package main

import (
	"aggregators"
	"fmt"
	// "regexp"
)

func main() {
	fmt.Println("Hi")
	agg := aggregators.NewCerberusAggregator("doralon")
	err := agg.Aggregate("")
	fmt.Println(err)
}

