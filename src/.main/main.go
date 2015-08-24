// Testing ground for aggregators.
package main

import (
	"fmt"
	"aggregators"
)

func main() {
	fmt.Println("Hi")
	fmt.Println(aggregators.Cerberus("freshmarket_sn", "f_efrd").Aggregate("try/freshmarket"))
}

