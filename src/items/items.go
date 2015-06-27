// Parses price XMLs.
package main

import (
	"regexp"
	"os"
	"io/ioutil"
	"fmt"
)

func main() {
	makeCaptures()
	pe("Parses price XMLs to TSV tables.")
	pe("Reading from stdin...")
	
	// Read input XML.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading input:", err)
		os.Exit(2)
	}
	
	// Parse fields.
	items := captureXml("Item").FindAllSubmatch(data, -1)
	pef("Found %d items.\n", len(items))
	for i := range items {
		item := items[i][1]
		for _, re := range fieldRes {
			field := re.FindSubmatch(item)
			if field != nil {
				fmt.Printf("%s", field[1])
			}
			fmt.Printf("\t")
		}
		fmt.Printf("\n")
	}
}

// Println to stderr.
func pe(a ...interface{}) {
	fmt.Println(a...)
}

// Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

var fields = []string {
	"PriceUpdateDate",
	"ItemCode",
	"ItemType",
	"ItemName",
	"ManufacturerName",
	"ManufacturerCountry",
	"ManufacturerItemDescription",
	"UnitQty",
	"Quantity",
	"UnitOfMeasure",
	"b(?:I|l)sWeighted",
	"QtyInPackage",
	"ItemPrice",
	"UnitOfMeasurePrice",
	"AllowDiscount",
	"ItemStatus",
}

var fieldRes []*regexp.Regexp

func captureXml(tag string) *regexp.Regexp {
	return regexp.MustCompile("(?si)<" + tag + ">(.*?)</" + tag + ">")
}

func makeCaptures() {
	fieldRes = make([]*regexp.Regexp, len(fields))
	for i := range fields {
		fieldRes[i] = captureXml(fields[i])
	}
}
