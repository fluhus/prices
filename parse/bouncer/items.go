package bouncer

// Handles reporting & bouncing of items.

import (
	"path/filepath"
	"sync"
)

var (
	itemsOut  *fileWriter // Output file.
	itemsLock sync.Mutex  // For synchronizing id generation.
	items     map[int]int // From item hash to item id.
)

// Initializes the 'items' table bouncer.
func initItems() {
	items = map[int]int{}
	if state.Items != nil {
		items = stringMapToIntMap(state.Items).(map[int]int)
	}

	var err error
	itemsOut, err = newTempFileWriter(filepath.Join(outDir, "items.txt"))
	if err != nil {
		panic(err)
	}
}

// Finalizes the 'items' table bouncer.
func finalizeItems() {
	itemsOut.Close()
	state.Items = intMapToStringMap(items).(map[string]int)
}

// A single entry in the 'items' table.
type Item struct {
	ItemType string
	ItemCode string
	ChainId  string
}

// Returns the hash of an item entry.
func (i *Item) hash() int {
	return hash(i.ItemType, i.ItemCode, i.ChainId)
}

// Returns a slice of item ids for the given items. Generates new ids if
// necessary. Thread safe.
func MakeItemIds(is []*Item) []int {
	itemsLock.Lock()
	defer itemsLock.Unlock()
	result := make([]int, len(is))
	for i := range is {
		result[i] = makeItemId(is[i])
	}
	return result
}

// Returns (and maybe generates) an id for the given item.
func makeItemId(i *Item) int {
	// Look up in hash table.
	h := i.hash()
	id, ok := items[h]
	if ok {
		return id
	}

	// Not found - assign new id and print it.
	id = len(items) + 1 // 1-based id's.
	items[h] = id

	itemsOut.printCsv(id, i.ItemType, i.ItemCode, i.ChainId)

	return id
}
