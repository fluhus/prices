package bouncer

// Handles reporting & bouncing of prices.

import (
	"fmt"
	"os"
	"bufio"
	"runtime"
	"path/filepath"
)

var pricesChan chan []*Price
var pricesDone chan int
// Maps itemId,storeId to hash.
var pricesMap map[int64]int
var pricesOut *os.File
var pricesOutBuf *bufio.Writer

func initPrices() {
	pricesChan = make(chan []*Price, runtime.NumCPU())
	pricesDone = make(chan int, 1)
	pricesMap = map[int64]int {}

	var err error
	pricesOut, err = os.Create(filepath.Join(outDir, "prices.txt"))
	if err != nil { panic(err) }
	pricesOutBuf = bufio.NewWriter(pricesOut)
	
	go func() {
		for prices := range pricesChan {
			reportPrices(prices)
		}
		pricesDone <- 0
	}()
}

func finalizePrices() {
	close(pricesChan)
	<-pricesDone
	pricesOutBuf.Flush()
	pricesOut.Close()
}

type Price struct {
	Timestamp int64
	ItemId int
	StoreId int
	Price string
	UnitOfMeasurePrice string
	UnitOfMeasure string
	Quantity string
};

func (p *Price) hash() int {
	return hash(
		p.Price,
		p.UnitOfMeasurePrice,
		p.UnitOfMeasure,
		p.Quantity,
	)
}

func (p *Price) id() int64 {
	return int64(p.ItemId) << 32 + int64(p.StoreId)
}

func ReportPrices(ps []*Price) {
	pricesChan <- ps
}

func reportPrices(ps []*Price) {
	for i := range ps {
		h := ps[i].hash()
		last := pricesMap[ps[i].id()]
		if h != last {
			pricesMap[ps[i].id()] = h
			fmt.Fprintf(pricesOutBuf, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
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



