package bouncer

// Handles reporting & bouncing of promos.

import (
	"encoding/json"
	"path/filepath"
	"runtime"
)

var (
	promosOut       *fileWriter        // Output file for 'promos'.
	promosItemsOut  *fileWriter        // Output file for 'promos_items'.
	promosStoresOut *fileWriter        // Output file for 'promos_stores'.
	promosToOut     *fileWriter        // Output file for 'promos_to'.
	promosChan      chan []*Promo      // Used for reporting promos.
	promosDone      chan int           // Indicates when promo reporting is finished.
	nextPromoId     int                // Id to assign to the next new promo.
	promosMap       map[int][]*promoId // Maps hash to promo-details.
)

// Initializes the 'promos*' table bouncer.
func initPromos() {
	// Initialize data structures.
	promosChan = make(chan []*Promo, runtime.NumCPU())
	promosDone = make(chan int, 1)
	nextPromoId = state.NextPromoId

	promosMap = map[int][]*promoId{}
	if state.PromosMap != nil {
		promosMap = stringMapToIntMap(state.PromosMap).(map[int][]*promoId)
	}

	// Open output files.
	var err error
	promosOut, err = newTempFileWriter(
		filepath.Join(outDir, "promos.txt"))
	if err != nil {
		panic(err)
	}
	promosItemsOut, err = newTempFileWriter(
		filepath.Join(outDir, "promos_items.txt"))
	if err != nil {
		panic(err)
	}
	promosStoresOut, err = newTempFileWriter(
		filepath.Join(outDir, "promos_stores.txt"))
	if err != nil {
		panic(err)
	}
	promosToOut, err = newTempFileWriter(
		filepath.Join(outDir, "promos_to.txt"))
	if err != nil {
		panic(err)
	}

	// Listen on channel for incoming promos.
	go func() {
		for promos := range promosChan {
			reportPromos(promos)
		}
		promosDone <- 0
	}()
}

// Finalizes the 'promos*' table bouncer.
func finalizePromos() {
	close(promosChan)
	<-promosDone

	for _, pids := range promosMap {
		for _, pid := range pids {
			promosToOut.printCsv(pid.Id, pid.TimestampTo+60*60*24)
		}
	}

	promosOut.Close()
	promosItemsOut.Close()
	promosStoresOut.Close()
	promosToOut.Close()

	state.NextPromoId = nextPromoId
	state.PromosMap = intMapToStringMap(promosMap).(map[string][]*promoId)
}

// A single entry in the 'promos' table.
type Promo struct {
	Timestamp                 int64
	ChainId                   string
	PromotionId               string
	PromotionDescription      string
	PromotionStartDate        string
	PromotionStartHour        string
	PromotionEndDate          string
	PromotionEndHour          string
	RewardType                string
	AllowMultipleDiscounts    string
	MinQty                    string
	MaxQty                    string
	DiscountRate              string
	DiscountType              string
	MinPurchaseAmnt           string
	MinNoOfItemOffered        string
	PriceUpdateDate           string
	DiscountedPrice           string
	DiscountedPricePerMida    string
	AdditionalIsCoupn         string
	AdditionalGiftCount       string
	AdditionalIsTotal         string
	AdditionalMinBasketAmount string
	Remarks                   string
	StoreId                   int
	ItemIds                   []int
	GiftItems                 []string
}

// Returns the hash of an store-meta entry.
func (p *Promo) hash() int {
	return hash(
		p.PromotionDescription,
		p.PromotionStartDate,
		p.PromotionStartHour,
		p.PromotionEndDate,
		p.PromotionEndHour,
		p.RewardType,
		p.AllowMultipleDiscounts,
		p.MinQty,
		p.MaxQty,
		p.DiscountRate,
		p.DiscountType,
		p.MinPurchaseAmnt,
		p.MinNoOfItemOffered,
		p.PriceUpdateDate,
		p.DiscountedPrice,
		p.DiscountedPricePerMida,
		p.AdditionalIsCoupn,
		p.AdditionalGiftCount,
		p.AdditionalIsTotal,
		p.AdditionalMinBasketAmount,
		p.Remarks,
		p.ItemIds,
		p.GiftItems,
	)
}

// Reports the given promos.
func ReportPromos(ps []*Promo) {
	promosChan <- ps
}

// Holds data about the last promo with the specific identification details.
type promoId struct {
	Id          int              // ID given to promo.
	ChainId     string           // ID of the reporting chain.
	PromotionId string           // ID reported by the chain.
	TimestampTo int64            // Last timestamp the promo was seen.
	StoreIds    map[int]struct{} // Ids of stores that reported that promo.
}

// Like promoId but JSON compatible.
type promoIdMarshaler struct {
	Id          int
	ChainId     string
	PromotionId string
	TimestampTo int64
	StoreIds    []int
}

func (p *promoId) MarshalJSON() ([]byte, error) {
	marsh := &promoIdMarshaler{p.Id, p.ChainId, p.PromotionId, p.TimestampTo,
		make([]int, 0, len(p.StoreIds))}
	for id := range p.StoreIds {
		marsh.StoreIds = append(marsh.StoreIds, id)
	}
	return json.Marshal(marsh)
}

func (p *promoId) UnmarshalJSON(data []byte) error {
	marsh := &promoIdMarshaler{}
	err := json.Unmarshal(data, marsh)
	if err != nil {
		return err
	}
	p.Id = marsh.Id
	p.ChainId = marsh.ChainId
	p.PromotionId = marsh.PromotionId
	p.TimestampTo = marsh.TimestampTo
	p.StoreIds = map[int]struct{}{}
	for _, id := range marsh.StoreIds {
		p.StoreIds[id] = struct{}{}
	}
	return nil
}

// Returns the promo-id object that corresponds to the given details. Returns
// nil if not found.
func lastReportedPromo(hash int, chainId string, promotionId string) *promoId {
	candidates := promosMap[hash]

	for _, candidate := range candidates {
		if candidate.ChainId == chainId &&
			candidate.PromotionId == promotionId {
			return candidate
		}
	}

	return nil
}

// Reports the given promos. Called by the goroutine that listens on the
// channel.
func reportPromos(ps []*Promo) {
	for _, p := range ps {
		h := p.hash()
		last := lastReportedPromo(h, p.ChainId, p.PromotionId)

		if last == nil {
			// Assign new id.
			last = &promoId{nextPromoId, p.ChainId, p.PromotionId, p.Timestamp,
				map[int]struct{}{}}
			promosMap[h] = append(promosMap[h], last)
			nextPromoId++

			// Report in promos_items.
			notInPromosItems := "0"
			if len(p.ItemIds) > 100 {
				notInPromosItems = "1"
			} else {
				for i := range p.ItemIds {
					promosItemsOut.printCsv(last.Id, p.ItemIds[i],
						p.GiftItems[i])
				}
			}

			// Report new promo.
			promosOut.printCsv(
				last.Id,
				p.Timestamp,
				0,
				p.ChainId,
				p.PromotionId,
				p.PromotionDescription,
				p.PromotionStartDate,
				p.PromotionStartHour,
				p.PromotionEndDate,
				p.PromotionEndHour,
				p.RewardType,
				p.AllowMultipleDiscounts,
				p.MinQty,
				p.MaxQty,
				p.DiscountRate,
				p.DiscountType,
				p.MinPurchaseAmnt,
				p.MinNoOfItemOffered,
				p.PriceUpdateDate,
				p.DiscountedPrice,
				p.DiscountedPricePerMida,
				p.AdditionalIsCoupn,
				p.AdditionalGiftCount,
				p.AdditionalIsTotal,
				p.AdditionalMinBasketAmount,
				p.Remarks,
				len(p.ItemIds),
				notInPromosItems,
			)
		} else {
			last.TimestampTo = p.Timestamp
		}

		// Report in promos_stores.
		if _, ok := last.StoreIds[p.StoreId]; !ok {
			last.StoreIds[p.StoreId] = struct{}{}
			promosStoresOut.printCsv(last.Id, p.StoreId)
		}
	}
}
