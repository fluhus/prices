package main

// Capturer type for capturing field values in XMLs.

import (
	"regexp"
)

// Captures fields for parsing XML files.
type capturer struct {
	re *regexp.Regexp
	column string
}

// Returns the content of the matching field, or nil if not found.
// Result may need to be trimmed.
func (c *capturer) capture(text []byte) []byte {
	match := c.re.FindSubmatch(text)
	if match == nil {
		return nil
	} else {
		return match[1]
	}
}

// Returns the content of the matching repeated field. Results may need to be
// trimmed.
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



