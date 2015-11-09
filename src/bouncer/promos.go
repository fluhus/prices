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

	promosMap map[string]*promoHash // Maps ChainId,PromotionId to hash.
)

// Initializes the 'promos*' table bouncer.
func initPromos() {
	// Initialize data structures.
	promosChan = make(chan []*Promo, runtime.NumCPU())
	promosDone = make(chan int, 1)
	promosMap = map[string]*promoHash{}
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
	for _, ph := range promosMap {
		fmt.Fprintf(promosToOutBuf, "%v\t%v\n", ph.id, ph.timestampTo+60*60*24)
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

// Holds data about the last promo with the specific identification details.
type promoHash struct {
	hash        int   // Hash of the promo.
	id          int   // Id given to promo.
	timestampTo int64 // Last timestamp the promo was seen.
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

// Returns the identification string of the promo.
func (p *Promo) id() string {
	return fmt.Sprintf("%v,%v", p.ChainId, p.PromotionId)
}

// Reports the given promos.
func ReportPromos(ps []*Promo) {
	promosChan <- ps
}

// Reports the given promos. Called by the goroutine that listens on the channel.
func reportPromos(ps []*Promo) {
	for _, p := range ps {
		h := p.hash()
		last := promosMap[p.id()]

		if last == nil || last.hash != h {
			// Report last timestamp of previous.
			if last != nil {
				fmt.Fprintf(promosToOutBuf, "%v\t%v\n", last.id,
					last.timestampTo+60*60*24)
			}

			// Assign new id.
			promosMap[p.id()] = &promoHash{h, nextPromoId, p.Timestamp}
			nextPromoId++

			// Report in promos_items table. TODO
			inPromosItems := true

			// Report new promo.
			fmt.Fprintf(promosOutBuf,
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t" +
						"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
				nextPromoId-1,
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
				inPromosItems,
			)
		} else {
			last.timestampTo = p.Timestamp
		}
	}
}

