package bouncer

// Handles reporting & bouncing of stores.

import (
	"path/filepath"
)

var (
	storesOut  *fileWriter   // Output file.
	storeToken chan int      // Token for synchronizing id generation.
	stores     []*Store      // Reported stores.
	storesMap  map[int][]int // Store hash-index.
)

// Initializes the 'stores' table bouncer.
func initStores() {
	storeToken = make(chan int, 1)
	storesMap = map[int][]int{}
	storeToken <- 0
	stores = state.Stores
	storesMap = map[int][]int{}
	if state.StoresMap != nil {
		storesMap = stringMapToIntMap(state.StoresMap).(map[int][]int)
	}

	var err error
	storesOut, err = newTempFileWriter(filepath.Join(outDir, "stores.txt"))
	if err != nil {
		panic(err)
	}
}

// Finalizes the 'stores' table bouncer.
func finalizeStores() {
	storesOut.Close()
	state.Stores = stores
	state.StoresMap = intMapToStringMap(storesMap).(map[string][]int)
}

// A single entry in the 'stores' table.
type Store struct {
	ChainId         string
	SubchainId      string
	ReportedStoreId string
}

// Returns the hash of an store entry.
func (s *Store) hash() int {
	return hash(s.ChainId, s.SubchainId, s.ReportedStoreId)
}

// Returns true only if all fields are equal.
func (s *Store) equals(t *Store) bool {
	return s.ChainId == t.ChainId &&
		s.SubchainId == t.SubchainId &&
		s.ReportedStoreId == t.ReportedStoreId
}

// Returns a slice of store ids for the given stores. Generates new ids if
// necessary. Thread safe.
func MakeStoreIds(ss []*Store) []int {
	<-storeToken
	result := make([]int, len(ss))
	for i := range ss {
		result[i] = makeStoreId(ss[i])
	}
	storeToken <- 0
	return result
}

// Returns (and maybe generates) an id for the given store.
func makeStoreId(s *Store) int {
	// Look up in hash table.
	h := s.hash()
	candidates := storesMap[h]

	// Compare to candidates.
	for _, c := range candidates {
		if s.equals(stores[c]) {
			return c
		}
	}

	// Not found - assign new id and print it.
	result := len(stores)
	storesMap[h] = append(storesMap[h], result)
	stores = append(stores, s)

	printTsv(storesOut,
		result,
		s.ChainId,
		s.SubchainId,
		s.ReportedStoreId)

	return result
}
