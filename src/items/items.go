// Parses price XMLs.
package main

import (
	"regexp"
	"os"
	"io/ioutil"
	"fmt"
	"encoding/xml"
	"bytes"
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
		item := items[i][2]
		itemMap, err := parseItem(item)
		if err != nil {
			pe("Error:", err)
			os.Exit(2)
		}
		pe(itemMap["ItemName"])
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

func parseItem(item []byte) (map[string]string, error) {
	result := map[string]string{}
	
	// Handle mandatory fields.
	for _, field := range mandatoryFields {
		value := field.FindSubmatch(item)
		if value == nil || len(trim(value[2])) == 0 {
			return nil, fmt.Errorf("No value for mandatory field '%s'.", field)
		}
		result[string(value[1])] = string(trim(value[2]))
	}
	
	// Handle optional fields.
	for _, field := range optionalFields {
		value := field.FindSubmatch(item)
		if value == nil || len(trim(value[2])) == 0 {
			continue
		}
		result[string(value[1])] = string(trim(value[2]))
	}
	
	return result, nil
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

// Returns a regex that captures an XML node surrounded by the given tag.
// Capture group 1 is the tag, and capture group 2 is the content of the node
// without the surrounding tag.
func captureXml(tag string) *regexp.Regexp {
	return regexp.MustCompile("(?si)<(" + tag + ")>(.*?)</" + tag + ">")
}

func trim(b []byte) []byte {
	return bytes.Trim(b, " \t\n\r")
}


