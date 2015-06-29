// Parses price XMLs.
package main

import (
	"regexp"
	"os"
	"io/ioutil"
	"fmt"
	"encoding/xml"
)

func main() {
	pe("Parses price XMLs to TSV tables.")
	pe("Reading from stdin...")
	
	// Read input XML.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading input:", err)
		os.Exit(2)
	}

	// Check XML validity.
	err = xml.Unmarshal(data, &struct{}{})
	if err != nil {
		pe("Error parsing input:", err)
		os.Exit(2)
	}
	
	// Parse items.
	items := captureXml("Item").FindAllSubmatch(data, -1)
	pef("Found %d items.\n", len(items))
	for i := range items {
		item := items[i][1]
		for _, f := range mandatoryFields {
			value := f.FindSubmatch(item)
			if value == nil || len(value[1]) == 0 {
				pef("Error: No value for mandatory field '%s'.\n", f)
			}
		}
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

var mandatoryFields = []*regexp.Regexp {
	captureXml("PriceUpdateDate"),
	captureXml("ItemCode"),
	captureXml("ItemName"),
	captureXml("ItemPrice"),
}

var optionalFields = []*regexp.Regexp {
	captureXml("ManufacturerName"),
	captureXml("ManufacturerCountry"),
	captureXml("ManufacturerItemDescription"),
	captureXml("UnitQty"),
	captureXml("Quantity"),
	captureXml("UnitOfMeasure"),
	captureXml("b(?:I|l)sWeighted"),
	captureXml("QtyInPackage"),
	captureXml("UnitOfMeasurePrice"),
	captureXml("AllowDiscount"),
	captureXml("ItemStatus"),
	captureXml("ItemType"),
}

func captureXml(tag string) *regexp.Regexp {
	return regexp.MustCompile("(?si)<" + tag + ">(.*?)</" + tag + ">")
}




