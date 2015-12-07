package main

// Capturer type, for looking up field values in XML nodes.

import (
	"github.com/fluhus/gostuff/xmlnode"
	"strings"
)

// Looks up field values in XML nodes.
type capturer struct {
	column string
	tags   []string
}

// Returns all nodes (including input) whos tag equals one of this capturer's
// tags.
func (c *capturer) findNodes(node xmlnode.Node) []xmlnode.Node {
	nodes := &[]xmlnode.Node{}
	c.findNodesRec(node, nodes)
	return *nodes
}

// Recursive function for findNodes.
func (c *capturer) findNodesRec(node xmlnode.Node, nodes *[]xmlnode.Node) {
	// Check if current node matches.
	for _, tag := range c.tags {
		if node.TagName() == tag {
			*nodes = append(*nodes, node)
			break
		}
	}

	// Search in children.
	for _, child := range node.Children() {
		c.findNodesRec(child, nodes)
	}
}

// Returns the text value of the first node whos tag matches one of this
// capturer's tags. The boolean return value indicates if such a node was found.
func (c *capturer) findValue(node xmlnode.Node) (string, bool) {
	// Check if current node matches.
	for _, tag := range c.tags {
		if node.TagName() == tag {
			if len(node.Children()) > 0 {
				return node.Children()[0].Text(), true
			} else {
				return "", true
			}
		}
	}

	// Search in children.
	for _, child := range node.Children() {
		text, ok := c.findValue(child)
		if ok {
			return text, true
		}
	}

	// Not found. :(
	return "", false
}

// Returns all nodes (including input) whos tag equals one of this capturer's
// tags.
func (c *capturer) findValues(node xmlnode.Node) []string {
	values := &[]string{}
	c.findValuesRec(node, values)
	return *values
}

// Recursive function for findValues.
func (c *capturer) findValuesRec(node xmlnode.Node, values *[]string) {
	// Check if current node matches.
	for _, tag := range c.tags {
		if node.TagName() == tag {
			if len(node.Children()) > 0 {
				*values = append(*values, node.Children()[0].Text())
			} else {
				*values = append(*values, "")
			}
			break
		}
	}

	// Search in children.
	for _, child := range node.Children() {
		c.findValuesRec(child, values)
	}
}

// Returns a capturer that captures XML nodes/values under the given tags.
// All values will be lower cased.
func newCapturer(column string, tags ...string) *capturer {
	newColumn := strings.ToLower(column)
	newTags := make([]string, len(tags))
	for i := range tags {
		newTags[i] = strings.ToLower(tags[i])
	}

	return &capturer{
		newColumn,
		newTags,
	}
}

// Returns a slice capturers, according to the given strings.
// Strings that begin with a colon (:) indicate a column name, and all the rest
// are tag names. Tags are associated with the last encountered column name.
func newCapturers(colsTags ...string) []*capturer {
	// Empty slices are nothing but trouble.
	if len(colsTags) == 0 {
		return nil
	}

	// Check that first element is a column name.
	if !strings.HasPrefix(colsTags[0], ":") {
		panic("First element must be a column name (begin with a colon).")
	}

	result := []*capturer{}
	lastColumn := 0

	for i, s := range colsTags {
		if strings.HasPrefix(s, ":") && i > 0 {
			result = append(result, newCapturer(
				colsTags[lastColumn][1:],
				colsTags[lastColumn+1:i]...,
			))

			lastColumn = i
		}
	}

	// Add last element.
	result = append(result, newCapturer(
		colsTags[lastColumn][1:],
		colsTags[lastColumn+1:len(colsTags)]...,
	))

	return result
}
