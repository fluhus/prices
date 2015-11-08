package main

import (
	"fmt"
	"os"
	"bufio"
	"runtime"
	"path/filepath"
)

func initItemMeta() {
	var err error
	itemMetaOut, err = os.Create(filepath.Join(args.outDir, "items_meta.txt"))
	if err != nil { panic(err) }
	itemMetaOutBuf = bufio.NewWriter(itemMetaOut)
	
	go func() {
		for metas := range itemMetaChan {
			reportItemMetas(metas)
		}
		itemMetaDone <- 0
	}()
}

var itemMetaChan = make(chan []*itemMeta, runtime.NumCPU())
var itemMetaDone = make(chan int, 1)

func finalizeItemMeta() {
	close(itemMetaChan)
	<-itemMetaDone
	itemMetaOutBuf.Flush()
	itemMetaOut.Close()
}

var itemMetaOut *os.File
var itemMetaOutBuf *bufio.Writer

type itemMeta struct {
	timestamp int64
	itemId int
	storeId int
	updateTime string
	itemName string
	manufacturerItemDescription string
	unitQuantity string
	isWeighted string
	quantityInPackage string
	allowDiscount string
	itemStatus string
};

func (i *itemMeta) hash() int {
	return hash(
		i.itemName,
		i.manufacturerItemDescription,
		i.unitQuantity,
		i.isWeighted,
		i.quantityInPackage,
		i.allowDiscount,
		i.itemStatus,
	)
}

func (i *itemMeta) id() int64 {
	return int64(i.itemId) << 32 + int64(i.storeId)
}

// Maps itemId,chainId to hash.
var itemMetaMap = map[int64]int {}

func reportItemMetas(is []*itemMeta) {
	for i := range is {
		h := is[i].hash()
		last := itemMetaMap[is[i].id()]
		if h != last {
			itemMetaMap[is[i].id()] = h
			fmt.Fprintf(itemMetaOutBuf, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
					is[i].timestamp,
					is[i].itemId,
					is[i].storeId,
					is[i].updateTime,
					is[i].itemName,
					is[i].manufacturerItemDescription,
					is[i].unitQuantity,
					is[i].isWeighted,
					is[i].quantityInPackage,
					is[i].allowDiscount,
					is[i].itemStatus,
			)
		}
	}
}



