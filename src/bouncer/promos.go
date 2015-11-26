package bouncer

// Handles reporting & bouncing of promos.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	promosOut          *os.File      // Output file for 'promos'.
	promosOutBuf       *bufio.Writer // Output buffer for 'promos'.
	promosItemsOut     *os.File      // Output file for 'promos_items'.
	promosItemsOutBuf  *bufio.Writer // Output buffer for 'promos_items'.
	promosStoresOut    *os.File      // Output file for 'promos_stores'.
	promosStoresOutBuf *bufio.Writer // Output buffer for 'promos_stores'.
	promosToOut        *os.File      // Output file for 'promos_to'.
	promosToOutBuf     *bufio.Writer // Output buffer for 'promos_to'.
	promosChan         chan []*Promo // Used for reporting promos.
	promosDone         chan int      // Indicates when promo reporting is finished.
	nextPromoId        int           // Id to assign to the next new promo.

	promosMap map[int][]*promoId // Maps hash to promo-details.
)

// Initializes the 'promos*' table bouncer.
func initPromos() {
	// Initialize data structures.
	promosChan = make(chan []*Promo, runtime.NumCPU())
	promosDone = make(chan int, 1)
	promosMap = map[int][]*promoId{}
	nextPromoId = 0

	// Open output files.
	var err error
	promosOut, err = os.Create(filepath.Join(outDir, "promos.txt"))
	if err != nil {
		panic(err)
	}
	promosOutBuf = bufio.NewWriter(promosOut)

	promosItemsOut, err = os.Create(filepath.Join(outDir, "promos_items.txt"))
	if err != nil {
		panic(err)
	}
	promosItemsOutBuf = bufio.NewWriter(promosItemsOut)

	promosStoresOut, err = os.Create(filepath.Join(outDir, "promos_stores.txt"))
	if err != nil {
		panic(err)
	}
	promosStoresOutBuf = bufio.NewWriter(promosStoresOut)

	promosToOut, err = os.Create(filepath.Join(outDir, "promos_to.txt"))
	if err != nil {
		panic(err)
	}
	promosToOutBuf = bufio.NewWriter(promosToOut)

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
	// Close channels.
	close(promosChan)
	<-promosDone

	// Write pending data.
	for _, pids := range promosMap {
		for _, pid := range pids {
			fmt.Fprintf(promosToOutBuf, "%v\t%v\n", pid.id,
				pid.timestampTo+60*60*24)
		}
	}

	// Flush output buffers and close files.
	promosOutBuf.Flush()
	promosOut.Close()
	promosItemsOutBuf.Flush()
	promosItemsOut.Close()
	promosStoresOutBuf.Flush()
	promosStoresOut.Close()
	promosToOutBuf.Flush()
	promosToOut.Close()
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
	id          int              // ID given to promo.
	chainId     string           // ID of the reporting chain.
	promotionId string           // ID reported by the chain.
	timestampTo int64            // Last timestamp the promo was seen.
	storeIds    map[int]struct{} // Ids of stores that reported that promo.
}

// Returns the promo-id object that corresponds to the given details. Returns
// nil if not found.
func lastReportedPromo(hash int, chainId string, promotionId string) *promoId {
	candidates := promosMap[hash]

	for _, candidate := range candidates {
		if candidate.chainId == chainId &&
			candidate.promotionId == promotionId {
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
					fmt.Fprintf(promosItemsOutBuf, "%v\t%v\t%v\n", last.id,
						p.ItemIds[i], p.GiftItems[i])
				}
			}

			// Report new promo.
			fmt.Fprintf(promosOutBuf,
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t"+
					"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
				last.id,
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
			last.timestampTo = p.Timestamp
		}

		// Report in promos_stores.
		if _, ok := last.storeIds[p.StoreId]; !ok {
			last.storeIds[p.StoreId] = struct{}{}
			fmt.Fprintf(promosStoresOutBuf, "%v\t%v\n", last.id, p.StoreId)
		}
	}
}
