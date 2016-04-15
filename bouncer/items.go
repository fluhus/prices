package bouncer

// Handles reporting & bouncing of items.

import (
	"path/filepath"
)

var (
	itemsOut   *fileWriter   // Output file.
	itemsToken chan int      // Token for synchronizing id generation.
	items      []*Item       // Reported items.
	itemsMap   map[int][]int // Item hash-index.
)

// Initializes the 'items' table bouncer.
func initItems() {
	itemsToken = make(chan int, 1)
	itemsToken <- 0

	items = state.Items
	itemsMap = map[int][]int{}
	if state.ItemsMap != nil {
		itemsMap = stringMapToIntMap(state.ItemsMap).(map[int][]int)
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
	state.Items = items
	state.ItemsMap = intMapToStringMap(itemsMap).(map[string][]int)
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

// Returns true only if all fields are equal.
func (i *Item) equals(j *Item) bool {
	return i.ItemType == j.ItemType &&
		i.ItemCode == j.ItemCode &&
		i.ChainId == j.ChainId
}

// Returns a slice of item ids for the given items. Generates new ids if
// necessary. Thread safe.
func MakeItemIds(is []*Item) []int {
	<-itemsToken
	result := make([]int, len(is))
	for i := range is {
		result[i] = makeItemId(is[i])
	}
	itemsToken <- 0
	return result
}

// Returns (and maybe generates) an id for the given item.
func makeItemId(i *Item) int {
	// Look up in hash table.
	h := i.hash()
	candidates := itemsMap[h]

	// Compare to candidates.
	for _, c := range candidates {
		if i.equals(items[c]) {
			return c
		}
	}

	// Not found - assign new id and print it.
	result := len(items)
	itemsMap[h] = append(itemsMap[h], result)
	items = append(items, i)

	printTsv(itemsOut, result, i.ItemType, i.ItemCode, i.ChainId)

	return result
}
