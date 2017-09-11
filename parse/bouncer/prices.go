package bouncer

// Handles reporting & bouncing of prices.

import (
	"path/filepath"
	"runtime"
)

var (
	pricesOut  *fileWriter   // Output file.
	pricesChan chan []*Price // Used for reporting prices.
	pricesDone chan int      // Indicates that price reporting is done.
	pricesMap  map[int]int   // Maps itemId,storeId to hash.
)

// Initializes the 'prices' table bouncer.
func initPrices() {
	pricesChan = make(chan []*Price, runtime.NumCPU())
	pricesDone = make(chan int, 1)

	pricesMap = map[int]int{}
	if state.PricesMap != nil {
		pricesMap = stringMapToIntMap(state.PricesMap).(map[int]int)
	}

	var err error
	pricesOut, err = newTempFileWriter(filepath.Join(outDir, "prices.txt"))
	if err != nil {
		panic(err)
	}

	go func() {
		for prices := range pricesChan {
			reportPrices(prices)
		}
		pricesDone <- 0
	}()
}

// Finalizes the 'prices' table bouncer.
func finalizePrices() {
	close(pricesChan)
	<-pricesDone
	pricesOut.Close()
	state.PricesMap = intMapToStringMap(pricesMap).(map[string]int)
}

// A single entry in the 'prices' table.
type Price struct {
	Timestamp          int64
	ItemId             int
	StoreId            int
	Price              string
	UnitOfMeasurePrice string
	UnitOfMeasure      string
	Quantity           string
}

// Returns the hash of a price entry.
func (p *Price) hash() int {
	return hash(
		p.Price,
		p.UnitOfMeasurePrice,
		p.UnitOfMeasure,
		p.Quantity,
	)
}

// Returns the identifier of an price entry, by item-id and store-id.
func (p *Price) id() int {
	return p.ItemId<<32 + p.StoreId
}

// Reports the given prices.
func ReportPrices(ps []*Price) {
	pricesChan <- ps
}

// Reports the given prices. Called by the goroutine that listens on the
// channel.
func reportPrices(ps []*Price) {
	for i := range ps {
		h := ps[i].hash()
		last := pricesMap[ps[i].id()]
		if h != last {
			pricesMap[ps[i].id()] = h
			pricesOut.printCsv(
				ps[i].Timestamp,
				ps[i].ItemId,
				ps[i].StoreId,
				ps[i].Price,
				ps[i].UnitOfMeasurePrice,
				ps[i].UnitOfMeasure,
				ps[i].Quantity,
			)
		}
	}
}
