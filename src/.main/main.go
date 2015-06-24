// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	fmt.Println("Hi")
	fmt.Println(aggregators.NewCerberusAggregator("doralon").Aggregate("/cs/grad/amitlavon/icore/try"))
}

