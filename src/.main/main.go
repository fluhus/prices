package main

import (
	"aggregators"
	"fmt"
	"time"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Hi")
	t := time.Now()
	agg := aggregators.NewMegaAggregator()
	err := agg.Aggregate("./files")
	if err != nil { fmt.Println(err) } else { fmt.Println("no error") }
	fmt.Println("took", time.Now().Sub(t))
}

