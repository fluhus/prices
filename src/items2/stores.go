package main

import (
	"fmt"
	"os"
	"bufio"
	"path/filepath"
)

func initStores() {
	storeToken <- 0

	var err error
	storesOut, err = os.Create(filepath.Join(args.outDir, "stores.txt"))
	if err != nil { panic(err) }
	storesOutBuf = bufio.NewWriter(storesOut)
}

func finalizeStores() {
	storesOutBuf.Flush()
	storesOut.Close()
}

var storesOut *os.File
var storesOutBuf *bufio.Writer

type store struct {
	chainId string
	subchainId string
	reportedStoreId string
}

func (s *store) hash() int {
	return hash(s.chainId, s.subchainId, s.reportedStoreId)
}

func (s *store) equals(t *store) bool {
	return s.chainId == t.chainId &&
		s.subchainId == t.subchainId &&
		s.reportedStoreId == t.reportedStoreId
}

func makeStoreIds(ss []*store) []int {
	<- storeToken
	result := make([]int, len(ss))
	for i := range ss {
		result[i] = makeStoreId(ss[i])
	}
	storeToken <- 0
	return result
}

func makeStoreId(s *store) int {
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

	fmt.Fprintf(storesOutBuf, "%d\t%s\t%s\t%s\n", result + 1, s.chainId,
			s.subchainId, s.reportedStoreId)

	return result
}

var storeToken = make(chan int, 1)
var stores []*store
var storesMap = map[int][]int {}


