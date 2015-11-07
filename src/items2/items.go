package main

import (
	"fmt"
	"hash/crc64"
	"os"
	"bufio"
	"path/filepath"
)

func initItems() {
	itemToken <- 0

	var err error
	itemsOut, err = os.Create(filepath.Join(args.outDir, "items.txt"))
	if err != nil { panic(err) }
	itemsOutBuf = bufio.NewWriter(itemsOut)
}

func finalizeItems() {
	itemsOutBuf.Flush()
	itemsOut.Close()
}

var itemsOut *os.File
var itemsOutBuf *bufio.Writer

type item struct {
	itemType string
	itemCode string
	chainId  string
}

func (i *item) hash() int {
	return hash(i.itemType, i.itemCode, i.chainId)
}

func (i *item) equals(j *item) bool {
	return i.itemType == j.itemType &&
			i.itemCode == j.itemCode &&
			i.chainId == j.chainId
}

func makeItemIds(is []*item) []int {
	<- itemToken
	result := make([]int, len(is))
	for i := range is {
		result[i] = makeItemId(is[i])
	}
	itemToken <- 0
	return result
}

func makeItemId(i *item) int {
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

	fmt.Fprintf(itemsOutBuf, "%d\t%s\t%s\t%s\n", result + 1, i.itemType,
		i.itemCode, i.chainId)
	
	return result
}

var itemToken = make(chan int, 1)
var items []*item
var itemsMap = map[int][]int {}

func hash(a interface{}, b ...interface{}) int {
	crc := crc64.New(crcTable)
	
	fmt.Fprint(crc, a)
	for _, element := range b {
		fmt.Fprint(crc, element)
	}
	
	return int(crc.Sum64())
}

