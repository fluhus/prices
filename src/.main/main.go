// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	fmt.Println("Hi")
	fmt.Println(aggregators.NewNibitAggregator(aggregators.Hashook, 2).Aggregate("try"))
}

