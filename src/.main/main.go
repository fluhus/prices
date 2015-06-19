package main

import (
	"aggregators"
	"fmt"
)

func main() {
	fmt.Println("Hi")
	agg := aggregators.NewCerberusAggregator("doralon")
	err := agg.Aggregate("./files")
	fmt.Println(err)
}

