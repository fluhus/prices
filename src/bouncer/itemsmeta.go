package bouncer

// Handles reporting & bouncing of item metadata.

import (
	"fmt"
	"os"
	"bufio"
	"runtime"
	"path/filepath"
)

var itemMetaChan chan []*ItemMeta
var itemMetaDone chan int
var itemMetaOut *os.File
var itemMetaOutBuf *bufio.Writer
// Maps itemId,chainId to hash.
var itemMetaMap map[int64]int

func initItemsMeta() {
	itemMetaChan = make(chan []*ItemMeta, runtime.NumCPU())
	itemMetaDone = make(chan int, 1)
	itemMetaMap = map[int64]int {}

	var err error
	itemMetaOut, err = os.Create(filepath.Join(outDir, "items_meta.txt"))
	if err != nil { panic(err) }
	itemMetaOutBuf = bufio.NewWriter(itemMetaOut)
	
	go func() {
		for metas := range itemMetaChan {
			reportItemMetas(metas)
		}
		itemMetaDone <- 0
	}()
}

func finalizeItemsMeta() {
	close(itemMetaChan)
	<-itemMetaDone
	itemMetaOutBuf.Flush()
	itemMetaOut.Close()
}

type ItemMeta struct {
	Timestamp int64
	ItemId int
	StoreId int
	UpdateTime string
	ItemName string
	ManufacturerItemDescription string
	UnitQuantity string
	IsWeighted string
	QuantityInPackage string
	AllowDiscount string
	ItemStatus string
};

func (i *ItemMeta) hash() int {
	return hash(
		i.ItemName,
		i.ManufacturerItemDescription,
		i.UnitQuantity,
		i.IsWeighted,
		i.QuantityInPackage,
		i.AllowDiscount,
		i.ItemStatus,
	)
}

func (i *ItemMeta) id() int64 {
	return int64(i.ItemId) << 32 + int64(i.StoreId)
}

func ReportItemMetas(is []*ItemMeta) {
	itemMetaChan <- is
}

func reportItemMetas(is []*ItemMeta) {
	for i := range is {
		h := is[i].hash()
		last := itemMetaMap[is[i].id()]
		if h != last {
			itemMetaMap[is[i].id()] = h
			fmt.Fprintf(itemMetaOutBuf, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
					is[i].Timestamp,
					is[i].ItemId,
					is[i].StoreId,
					is[i].UpdateTime,
					is[i].ItemName,
					is[i].ManufacturerItemDescription,
					is[i].UnitQuantity,
					is[i].IsWeighted,
					is[i].QuantityInPackage,
					is[i].AllowDiscount,
					is[i].ItemStatus,
			)
		}
	}
}



