package main

// Parser type for parsing XMLs of prices, stores and promos.

import (
	"fmt"
	"bytes"
	"encoding/xml"
)

// Parses entire files and returns entries ready to be fed into a table.
type parser struct {
	// Capturer for dividing the file into items.
	divider         *capturer

	// Fields that may appear once per file. All are mandatory.
	globalFields    []*capturer
	
	// Mandatory fields that appear on every item.
	mandatoryFields []*capturer
	
	// Optional fields that may appear on every item.
	optionalFields  []*capturer
	
	// Repeated fields that may appear on every item.
	repeatedFields  []*capturer
}

// Returns a map for each item in the given text. Each map contains all columns,
// even those with no values. Returns an error if a mandatory value is missing.
func (p *parser) parse(text []byte) ([]map[string]string, error) {
	// Check XML validity.
	err := xml.Unmarshal(text, &struct{}{})
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	
	// Initialize result.
	items := p.divider.captures(text)
	result := make([]map[string]string, len(items))
	
	// Handle global fields.
	globals := toMap(p.globalFields, text)
	err = findMissing(globals)
	if err != nil {
		return nil, err
	}
	
	// Parse items.
	for i := range items {
		// Handle mandatory fields.
		mandatory := toMap(p.mandatoryFields, items[i])
		err = findMissing(mandatory)
		if err != nil {
			return nil, err
		}
		
		// Handle optional & repeated fields.
		optional := toMap(p.optionalFields, items[i])
		repeated := toMapRepeated(p.repeatedFields, items[i])
		
		result[i] = join(globals, mandatory, optional, repeated)
	}
	
	return result, nil
}

// Generates a map from column name to trimmed value, for each capturer.
func toMap(c []*capturer, text []byte) map[string]string {
	result := map[string]string {}
	for i := range c {
		result[c[i].column] = string(trim(c[i].capture(text)))
	}
	return result
}

// Generates a map from column name to trimmed repeated values, for each
// capturer. Repeated values are stored in a single string, separated by ';'.
func toMapRepeated(c []*capturer, text []byte) map[string]string {
	result := map[string]string {}
	for i := range c {
		values := c[i].captures(text)
		for _, value := range values {
			if result[c[i].column] == "" {
				result[c[i].column] = string(trim(value))
			} else {
				result[c[i].column] += ";" + string(trim(value))
			}
		}
	}
	return result
}

// Generates an error that reports missing fields in the map. Returns nil if no
// fields are missing.
func findMissing(m map[string]string) error {
	err := ""
	
	for field := range m {
		if m[field] == "" {
			if len(err) == 0 {
				err += field
			} else {
				err += ", " + field
			}
		}
	}
	
	if len(err) == 0 {
		return nil
	} else {
		return fmt.Errorf("Missing fields: " + err)
	}
}

// Trims whitespaces from a given byte array.
func trim(b []byte) []byte {
	return bytes.Trim(b, " \t\n\r")
}

// Returns a unified map that contains the entries from all maps. Overlaps
// will have the last value encountered. The input maps are unchanged.
func join(m ...map[string]string) map[string]string {
	result := map[string]string {}
	for i := range m {
		for j := range m[i] {
			result[j] = m[i][j]
		}
	}
	return result
}

