package bouncer

// Handles reporting & bouncing of stores.

import (
	"fmt"
	"os"
	"bufio"
	"path/filepath"
)

var storeToken chan int
var stores []*Store
var storesMap map[int][]int
var storesOut *os.File
var storesOutBuf *bufio.Writer

func initStores() {
	storeToken = make(chan int, 1)
	storesMap = map[int][]int {}
	storeToken <- 0

	var err error
	storesOut, err = os.Create(filepath.Join(outDir, "stores.txt"))
	if err != nil { panic(err) }
	storesOutBuf = bufio.NewWriter(storesOut)
}

func finalizeStores() {
	storesOutBuf.Flush()
	storesOut.Close()
}

type Store struct {
	ChainId string
	SubchainId string
	ReportedStoreId string
}

func (s *Store) hash() int {
	return hash(s.ChainId, s.SubchainId, s.ReportedStoreId)
}

func (s *Store) equals(t *Store) bool {
	return s.ChainId == t.ChainId &&
		s.SubchainId == t.SubchainId &&
		s.ReportedStoreId == t.ReportedStoreId
}

func MakeStoreIds(ss []*Store) []int {
	<- storeToken
	result := make([]int, len(ss))
	for i := range ss {
		result[i] = makeStoreId(ss[i])
	}
	storeToken <- 0
	return result
}

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

	fmt.Fprintf(storesOutBuf, "%d\t%s\t%s\t%s\n",
			result + 1,
			s.ChainId,
			s.SubchainId,
			s.ReportedStoreId)

	return result
}




