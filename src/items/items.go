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
	//pe("Parses price XMLs to TSV tables.")
	//pe("Reading from stdin...")
	
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
	
	// Get global fields.
	globals, err := parseGlobalFields(data)
	if err != nil {
		pe("Error:", err)
		os.Exit(2)
	}
	
	// Parse items.
	items := newXmlCapturer("(?:Item|Product)", "").captures(data)
	if len(items) == 0 {
		pe("Error: Found 0 items.")
		os.Exit(2)
	}
	
	for _, item := range items {
		itemMap, err := parseItem(item)
		if err != nil {
			pe("Error:", err)
			os.Exit(2)
		}
		itemMap = join(itemMap, globals)
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

func parseGlobalFields(data []byte) (map[string]string, error) {
	result := map[string]string {}
	for _, field := range globalFields {
		value := trim(field.capture(data))
		if len(value) == 0 {
			return nil, fmt.Errorf("No value for mandatory field '%s'.",
					field.column)
		}
		result[field.column] = string(value)
	}
	return result, nil
}

func parseItem(item []byte) (map[string]string, error) {
	result := map[string]string{}
	
	// Handle mandatory fields.
	for _, field := range mandatoryFields {
		value := trim(field.capture(item))
		if len(value) == 0 {
			return nil, fmt.Errorf("No value for mandatory field '%s'.",
					field.column)
		}
		result[field.column] = string(value)
	}
	
	// Handle optional fields.
	for _, field := range optionalFields {
		value := trim(field.capture(item))
		result[field.column] = string(value)
	}
	
	return result, nil
}

var globalFields = newXmlCapturers(
	"ChainId", "chain_id",
	"SubchainId", "subchain_id",
	"StoreId", "store_id",
)

var mandatoryFields = newXmlCapturers(
	"PriceUpdateDate", "update_time",
	"ItemCode", "item_id", 
	"ItemName", "item_name", 
	"ItemPrice", "price",
)

var optionalFields = newXmlCapturers(
	"ManufacturerName","manufacturer_name",
	"ManufacturerCountry","manufacturer_country",
	"ManufacturerItemDescription","manufacturer_item_description",
	"UnitQty","unit_quantity",
	"Quantity","quantity",
	"UnitOfMeasure","unit_of_measure",
	"b(?:I|l)sWeighted","is_weighted",
	"QtyInPackage","quantity_in_package",
	"UnitOfMeasurePrice","unit_of_measure_price",
	"AllowDiscount","allow_discount",
	"ItemStatus","item_status",
	"ItemType","item_type",
)

type capturer struct {
	re *regexp.Regexp
	column string
}

func (c *capturer) capture(text []byte) []byte {
	match := c.re.FindSubmatch(text)
	if match == nil {
		return nil
	} else {
		return match[1]
	}
}

func (c *capturer) captures(text []byte) [][]byte {
	match := c.re.FindAllSubmatch(text, -1)
	result := make([][]byte, len(match))
	
	for i := range match {
		result[i] = match[i][1]
	}
	
	return result
}

// Returns a capturer that captures an XML node surrounded by the given tag.
// Capture group 1 is the content of the node without the surrounding tag.
func newXmlCapturer(tag, column string) *capturer {
	return &capturer{
		regexp.MustCompile("(?si)<" + tag + ">(.*?)</" + tag + ">"),
		column}
}

// Returns a list of regexs created by captureXml, one for each input string.
func newXmlCapturers(tagsCols ...string) []*capturer {
	if len(tagsCols) % 2 == 1 {
		panic("Odd number of arguments is not accepted.")
	}

	result := make([]*capturer, len(tagsCols) / 2)
	for i := 0; i < len(tagsCols); i += 2 {
		result[i/2] = newXmlCapturer(tagsCols[i], tagsCols[i+1])
	}
	
	return result
}

func trim(b []byte) []byte {
	return bytes.Trim(b, " \t\n\r")
}

func join(a, b map[string]string) map[string]string {
	result := map[string]string {}
	for i := range a {
		result[i] = a[i]
	}
	for i := range b {
		result[i] = b[i]
	}
	return result
}

