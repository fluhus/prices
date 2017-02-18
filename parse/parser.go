package main

// Parser type for converting XML text data to field maps.

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	
	"github.com/fluhus/gostuff/xmlnode"
)

// Version of the parser, to detect data that were parsed with an outdated
// parser so they can be reprocessed. For readability and uniqueness, the
// version is the time when the parser was last modified.
const parserVersion = "23/3/2016 17:00"

// Parses entire XML files and returns maps that map each required field to its
// value.
type parser struct {
	// Capturer for dividing the file into items.
	divider *capturer

	// Fields that may appear once per file. All are mandatory.
	globalFields []*capturer

	// Mandatory fields that appear on every item.
	mandatoryFields []*capturer

	// Optional fields that may appear on every item.
	optionalFields []*capturer

	// Repeated fields that may appear on every item.
	repeatedFields []*capturer
}

// Returns a map for each item in the given text. Each map contains all columns,
// even those with no values. Returns an error if a mandatory value is missing.
// The preset argument contains preset values for fields, in case they are not
// found in the data. If they are found, the values in the data will be used.
func (p *parser) parse(text []byte, preset map[string]string) (
	[]map[string]string, error) {
	// Create XML node.
	node, err := xmlnode.ReadAll(bytes.NewBuffer(text))
	if err != nil {
		return nil, err
	}

	node = newLowercaseNode(node)

	// Initialize result.
	items := p.divider.findNodes(node)
	result := make([]map[string]string, len(items))

	// Handle global fields.
	globals := join(preset, toMap(p.globalFields, node))
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
func toMap(c []*capturer, node xmlnode.Node) map[string]string {
	result := map[string]string{}
	for i := range c {
		value, _ := c[i].findValue(node)
		result[c[i].column] = cleanFieldValue(value)
	}
	return result
}

// Generates a map from column name to trimmed repeated values, for each
// capturer. Repeated values are stored in a single string, separated by ';'.
func toMapRepeated(c []*capturer, node xmlnode.Node) map[string]string {
	result := map[string]string{}
	for i := range c {
		buf := make([]byte, 0)
		values := c[i].findValues(node)
		for _, value := range values {
			if len(buf) == 0 {
				buf = append(buf, cleanFieldValue(value)...)
			} else {
				buf = append(buf, ';')
				buf = append(buf, cleanFieldValue(value)...)
			}
		}
		result[c[i].column] = string(buf)
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

// Removes non visible ascii and non aleph-bet characters from the given string,
// replacing them with spaces. Each sequence of those will become a single
// whitespace.
// Also trims whitespaces from the beginning and end of the string.
func cleanFieldValue(s string) string {
	// TODO(amit): Move unreadable character handling to correctxml.go.
	return unreadableCharsRegexp.ReplaceAllString(strings.Trim(s, " \t\n\r"),
		" ")
}

// Characters to replace with a single space. Includes whitespaces and
// non-visible ascii/Hebrew characters.
var unreadableCharsRegexp = regexp.MustCompile("(\\s|[^ -~א-ת])+")

// Returns a unified map that contains the entries from all maps. Overlaps
// will have the last value encountered. Empty string values are ignored. The
// input maps are unchanged.
func join(m ...map[string]string) map[string]string {
	result := map[string]string{}
	for i := range m {
		for j := range m[i] {
			if result[j] == "" || m[i][j] != "" {
				result[j] = m[i][j]
			}
		}
	}
	return result
}

// An XML node with a lowercase tag name. Used to avoid recalculating lowercase
// tags for case-insensitive matching.
type lowercaseNode struct {
	xmlnode.Node
	tagName string
}

// Returns the given node tree wrapped by the LowercaseNode decorator. The input
// tree will be changed.
func newLowercaseNode(node xmlnode.Node) xmlnode.Node {
	children := node.Children()
	for i := range children {
		children[i] = newLowercaseNode(children[i])
	}

	lower := strings.ToLower(node.TagName())
	if node.TagName() == lower {
		return node
	}

	return &lowercaseNode{node, lower}
}

// Override TagName with lowercase variant.
func (n *lowercaseNode) TagName() string {
	return n.tagName
}
