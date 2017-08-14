package bouncer

// Handles reporting & bouncing of item metadata.

import (
	"path/filepath"
	"runtime"
)

var (
	itemMetaOut  *fileWriter      // Output file.
	itemMetaChan chan []*ItemMeta // Used for reporting item-metas.
	itemMetaDone chan int         // Indicates when meta reporting is finished.
	itemMetaMap  map[int]struct{} // Maps hash to item details.
)

// Initializes the 'items_meta' table bouncer.
func initItemsMeta() {
	itemMetaChan = make(chan []*ItemMeta, runtime.NumCPU())
	itemMetaDone = make(chan int, 1)

	itemMetaMap = map[int]struct{}{}
	if state.ItemMetaMap != nil {
		itemMetaMap = stringMapToIntMap(state.ItemMetaMap).(map[int]struct{})
	}

	var err error
	itemMetaOut, err =
		newTempFileWriter(filepath.Join(outDir, "items_meta.txt"))
	if err != nil {
		panic(err)
	}

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
	itemMetaOut.Close()
	state.ItemMetaMap =
		intMapToStringMap(itemMetaMap).(map[string]struct{})
}

// A single entry in the 'items_meta' table.
type ItemMeta struct {
	Timestamp                   int64
	ItemId                      int
	ChainId                     string
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

// Identifies a single hashed entry in the hash map.
type itemMetaId struct {
	ItemId  int
	ChainId string
}

// Reports the given metas.
func ReportItemMetas(is []*ItemMeta) {
	itemMetaChan <- is
}

// Reports the given metas. Called by the goroutine that listens on the channel.
func reportItemMetas(is []*ItemMeta) {
	for i := range is {
		h := is[i].hash()
		_, ok := itemMetaMap[h]
		if ok {
			return
		}
		itemMetaMap[h] = struct{}{}
		printTsv(itemMetaOut,
			is[i].Timestamp,
			is[i].ItemId,
			is[i].ChainId,
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
