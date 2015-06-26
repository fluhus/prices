// Parses price XMLs.
package main

import (
	"regexp"
	"os"
	"io/ioutil"
	"fmt"
)

func main() {
	pef("Parses price XMLs to TSV tables.\n")
	pef("Reading from stdin...\n")
	
	// Read input XML.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pef("Error reading input: %v\n", err)
		os.Exit(2)
	}
	
	// Parse fields.
	fields := []string{
		"Chainid",
		"SubChainid",
		"Storeid",
		"BikoretNo",
		"PriceUpdateDate",
		"PriceUpdateTime",
		"ItemCode",
		"ItemType",
		"ItemName",
		"ManufacturerName",
		"ManufacturerItemCountry",
		"ManufacturerItemDescription",
		"UnitQty",
		"Quantity",
		"UnitOfMeasure",
		"BlsWeighted",
		"QtyInPackage",
		"ItemPrice",
		"UnitOfMeasurePrice",
		"AllowDiscount",
		"ItemStatus",
	}
	
	match := fieldRegexp(fields).FindAllSubmatch(data, -1)

	// Print findings.
	for i := range fields {
		if i == 0 {
			fmt.Printf("%s", fields[i])
		} else {
			fmt.Printf("\t%s", fields[i])
		}
	}
	fmt.Printf("\n")

	for i := range match {
		for j := range match[i] {
			if j == 0 { continue }
			if j == 1 {
				fmt.Printf("%s", match[i][j])
			} else {
				fmt.Printf("\t%s", match[i][j])
			}
		}
		fmt.Printf("\n")
	}
}

// Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

func fieldRegexp(fields []string) *regexp.Regexp {
	re := "(?s)"
	for _, field := range fields {
		re += fmt.Sprintf(".*?<%s>(.*?)</%s>", field, field)
	}
	return regexp.MustCompile(re)
}
