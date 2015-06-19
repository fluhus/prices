package main

import (
	"aggregators"
	"fmt"
	"time"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// log.SetFlags(19)
	fmt.Println("Hi")
	t := time.Now()
	agg := aggregators.NewCerberusAggregator("osherad")
	err := agg.Aggregate("./files")
	if err != nil { fmt.Println(err) } else { fmt.Println("no error") }
	fmt.Println("took", time.Now().Sub(t))
}

