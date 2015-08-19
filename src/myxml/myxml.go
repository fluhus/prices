// Handles representation of XML structures.
package myxml

import (
	"encoding/xml"
	"regexp"
	"io"
)

// A single XML node. This type is the entry point to this package.
// Each XML tag and textual datum is represented by a node.
//
// Tag and text nodes are represented by the same type, to avoid
// over-complicating stuff with polymorphism.
type Node struct {
	Tag         string              // Name of the tag. Empty for text nodes.
	Attr        map[string]string   // Maps attribute name to value.
	Parent      *Node               // Parent node. Nil for head node.
	Children    []*Node             // Children nodes by order of appearance.
	                                // Nil for text nodes.
	IsText      bool                // Is this a text node.
	Text        string              // Textual data for text nodes. Empty for
	                                // non-text nodes.
}

// Reads all XML data from the given reader and stores it in a head node.
func Read(r io.Reader) (*Node, error) {
	// Create head node.
	result := &Node {
		"(head)",
		map[string]string {},
		nil,
		nil,
		false,
		"",
	}
	dec := xml.NewDecoder(r)
	
	var t xml.Token
	var err error
	current := result
	for t, err = dec.Token(); err == nil; t, err = dec.Token() {
		switch t := t.(type) {
		case xml.StartElement:
			// Create an attribute map.
			attrs := map[string]string {}
			for _, attr := range t.Attr {
				attrs[attr.Name.Local] = attr.Value
			}
			
			// Create child node.
			child := &Node {
				t.Name.Local,
				attrs,
				current,
				nil,
				false,
				"",
			}
			
			current.Children = append(current.Children, child)
			current = child
		
		case xml.EndElement:
			current = current.Parent
		
		case xml.CharData:
			child := &Node {
				"",
				nil,
				current,
				nil,
				true,
				string(t),
			}
			
			current.Children = append(current.Children, child)
		}
	}
	
	if err != io.EOF {
		return nil, err
	}
	
	return result, nil
}

// A simple string representation of a node, for debugging.
func (n *Node) String() string {
	return n.stringPrefix("")
}

// A recursive stringifier for nodes, for debugging.
func (n *Node) stringPrefix(prefix string) string {
	if n.IsText {
		return prefix + "Text: " + n.Text + "\n"
	} else {
		result := prefix + "Node: <" + n.Tag + ">"
		
		for attr, value := range n.Attr {
			result += " " + attr + "=\"" + value + "\""
		}
		result += "\n"
		
		for _, child := range n.Children {
			result += child.stringPrefix(prefix + "\t")
		}
		
		return result
	}
}

// Returns all nodes whos tags match the given regexp.
func (n *Node) FindTag(re *regexp.Regexp) []*Node {
	ns := &[]*Node{}
	n.findTagRec(re, ns)
	return *ns
}

// Recursively searches for nodes whos tags match the given regexp.
func (n *Node) findTagRec(re *regexp.Regexp, ns *[]*Node) {
	if n.IsText {
		return
	}
	
	if re.MatchString(n.Tag) {
		*ns = append(*ns, n)
	}
	
	for _, child := range n.Children {
		child.findTagRec(re, ns)
	}
}

func (n *Node) FindTextUnderTag(re *regexp.Regexp) (string, bool) {
	// If current node matches, return its text.
	if re.MatchString(n.Tag) {
		if len(n.Children) > 0 {
			return n.Children[0].Text, true
		} else {
			return "", true
		}
	}
	
	// Search in children.
	for _, child := range n.Children {
		text, ok := child.FindTextUnderTag(re)
		if ok {
			return text, true
		}
	}
	
	// Not found. :(
	return "", false
}

func (n *Node) FindAllTextUnderTag(re *regexp.Regexp) []string {
	ss := &[]string{}
	n.findAllTextUnderTagRec(re, ss)
	return *ss
}

func (n *Node) findAllTextUnderTagRec(re *regexp.Regexp, ss *[]string) {
	// If current node matches, append its text.
	if re.MatchString(n.Tag) {
		if len(n.Children) > 0 {
			*ss = append(*ss, n.Children[0].Text)
		} else {
			*ss = append(*ss, "")
		}
	}
	
	// Search in children.
	for _, child := range n.Children {
		child.findAllTextUnderTagRec(re, ss)
	}
}

