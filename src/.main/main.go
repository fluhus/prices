package main

import (
	"aggregators"
	"fmt"
)

func main() {
	agg := aggregators.NewCerberusAggregator("doralon")
	err := agg.Aggregate("")
	fmt.Println(err)
}

