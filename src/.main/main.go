// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	a := aggregators.NewNibitAggregator("")
	fmt.Println(a.Aggregate(""))
}

