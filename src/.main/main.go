// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	fmt.Println("Hi")
	fmt.Println(aggregators.Eden().Aggregate("/cs/grad/amitlavon/icore/try"))
}

