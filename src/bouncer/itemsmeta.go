package bouncer

// Handles reporting & bouncing of item metadata.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	itemMetaOut    *os.File         // Output file.
	itemMetaOutBuf *bufio.Writer    // Output buffer.
	itemMetaChan   chan []*ItemMeta // Used for reporting item-metas.
	itemMetaDone   chan int         // Indicates when meta reporting is 
	                                // finished.
	itemMetaMap    map[int64]int    // Maps ItemId & StoreId to hash.
)

// Initializes the 'items_meta' table bouncer.
func initItemsMeta() {
	itemMetaChan = make(chan []*ItemMeta, runtime.NumCPU())
	itemMetaDone = make(chan int, 1)
	itemMetaMap = map[int64]int{}

	var err error
	itemMetaOut, err = os.Create(filepath.Join(outDir, "items_meta.txt"))
	if err != nil {
		panic(err)
	}
	itemMetaOutBuf = bufio.NewWriter(itemMetaOut)

	go func() {
		for metas := range itemMetaChan {
			reportItemMetas(metas)
		}
		itemMetaDone <- 0
	}()
}

// Finalizes the 'items_meta' table bouncer.
func finalizeItemsMeta() {
	close(itemMetaChan)
	<-itemMetaDone
	itemMetaOutBuf.Flush()
	itemMetaOut.Close()
}

// A single entry in the 'items_meta' table.
type ItemMeta struct {
	Timestamp                   int64
	ItemId                      int
	StoreId                     int
	UpdateTime                  string
	ItemName                    string
	ManufacturerItemDescription string
	UnitQuantity                string
	IsWeighted                  string
	QuantityInPackage           string
	AllowDiscount               string
	ItemStatus                  string
}

// Returns the hash of an item-meta entry.
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

// Returns the identifier of an item-meta entry, by item-id and store-id.
func (i *ItemMeta) id() int64 {
	return int64(i.ItemId)<<32 + int64(i.StoreId)
}

// Reports the given metas.
func ReportItemMetas(is []*ItemMeta) {
	itemMetaChan <- is
}

// Reports the given metas. Called by the goroutine that listens on the channel.
func reportItemMetas(is []*ItemMeta) {
	for i := range is {
		h := is[i].hash()
		last := itemMetaMap[is[i].id()]
		if h != last {
			itemMetaMap[is[i].id()] = h
			fmt.Fprintf(itemMetaOutBuf,
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
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

