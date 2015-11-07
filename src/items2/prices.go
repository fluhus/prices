package main

import (
	"fmt"
	"os"
	"bufio"
	"runtime"
	"path/filepath"
)

func initPriceData() {
	var err error
	priceDataOut, err = os.Create(filepath.Join(args.outDir, "prices.txt"))
	if err != nil { panic(err) }
	priceDataOutBuf = bufio.NewWriter(priceDataOut)
	
	go func() {
		for prices := range priceDataChan {
			reportPriceData(prices)
		}
		priceDataDone <- 0
	}()
}

var priceDataChan = make(chan []*priceData, runtime.NumCPU())
var priceDataDone = make(chan int, 1)

func finalizePriceData() {
	close(priceDataChan)
	<-priceDataDone
	priceDataOutBuf.Flush()
	priceDataOut.Close()
}

var priceDataOut *os.File
var priceDataOutBuf *bufio.Writer

type priceData struct {
	timestamp int64
	itemId int
	storeId int
	price string
	unitOfMeasurePrice string
	unitOfMeasure string
	quantity string
};

func (p *priceData) hash() int {
	return hash(
		p.price,
		p.unitOfMeasurePrice,
		p.unitOfMeasure,
		p.quantity,
	)
}

func (p *priceData) id() int64 {
	return int64(p.itemId) << 32 + int64(p.storeId)
}

// Maps itemId,storeId to hash.
var priceDataMap = map[int64]int {}

func reportPriceData(ps []*priceData) {
	for i := range ps {
		h := ps[i].hash()
		last := priceDataMap[ps[i].id()]
		if h != last {
			priceDataMap[ps[i].id()] = h
			fmt.Fprintf(priceDataOutBuf, "%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
					ps[i].timestamp,
					ps[i].itemId,
					ps[i].storeId,
					ps[i].price,
					ps[i].unitOfMeasurePrice,
					ps[i].unitOfMeasure,
					ps[i].quantity,
			)
		}
	}
}



