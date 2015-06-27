// Parses price XMLs.
package main

import (
	"regexp"
	"os"
	"io/ioutil"
	"fmt"
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
	
	// Parse items.
	items := captureXml("Item").FindAllSubmatch(data, -1)
	pef("Found %d items.\n", len(items))
	for i := range items {
		item := items[i][1]
		for _, f := range fields {
			value := f.re().FindSubmatch(item)
			if f.mandatory && (value == nil || len(value[1]) == 0) {
				pef("Error: No value for mandatory field '%s'.", f.tag)
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

type field struct {
	tag string
	global bool
	mandatory bool
	rgx *regexp.Regexp
}

func (f field) re() *regexp.Regexp {
	if f.rgx == nil {
		f.rgx = captureXml(f.tag)
	}
	return f.rgx
}

var fields = []*field {
	&field{"PriceUpdateDate", false, true, nil},
	&field{"ItemCode", false, true, nil},
	&field{"ItemType", false, false, nil},
	&field{"ItemName", false, true, nil},
	&field{"ManufacturerName", false, false, nil},
	&field{"ManufacturerCountry", false, false, nil},
	&field{"ManufacturerItemDescription", false, false, nil},
	&field{"UnitQty", false, false, nil},
	&field{"Quantity", false, false, nil},
	&field{"UnitOfMeasure", false, false, nil},
	&field{"b(?:I|l)sWeighted", false, false, nil},
	&field{"QtyInPackage", false, false, nil},
	&field{"ItemPrice", false, true, nil},
	&field{"UnitOfMeasurePrice", false, false, nil},
	&field{"AllowDiscount", false, false, nil},
	&field{"ItemStatus", false, false, nil},
}

func captureXml(tag string) *regexp.Regexp {
	return regexp.MustCompile("(?si)<" + tag + ">(.*?)</" + tag + ">")
}




