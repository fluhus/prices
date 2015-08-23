// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	fmt.Println("Hi")
	fmt.Println(aggregators.Nibit(aggregators.Lahav, 2).Aggregate("nibit"))
}

