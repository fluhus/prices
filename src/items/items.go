// Parses price XMLs.
package main

import (
	"os"
	"io/ioutil"
	"fmt"
)

func main() {
	//pe("Parses price XMLs to TSV tables.")
	//pe("Reading from stdin...")
	
	// Read input XML.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading input:", err)
		os.Exit(2)
	}
	
	// Parse items.
	items, err := pricesParser.parse(data)
	pe(items, err)
}

// Println to stderr.
func pe(a ...interface{}) {
	fmt.Println(a...)
}

// Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}






