package main

// Functionality for correcting XML syntax and encoding errors.

import (
	"bytes"
	"regexp"
	"io/ioutil"
	"golang.org/x/net/html/charset"
)

// Converts the given XML to utf-8, and corrects some syntax errors that the
// publishers make.
func correctXml(text []byte) ([]byte, error) {
	// Charset reader converts arbitrary text to UTF-8. Hurray!
	r, err := charset.NewReader(bytes.NewBuffer(text), "application/xml")
	if err != nil {
		return nil, err
	}
	
	newText, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	// Some Gibberish will not convert, so convert manually.
	for i := byte(160); i <= 186; i++ {
		newText = bytes.Replace(newText,
				[]byte{195, i}, []byte{215, i-16}, -1)
	}
	
	// Replace encoding field with utf-8.
	newText = regexp.MustCompile("encoding=\".*?\"").ReplaceAll(newText,
			[]byte("encoding=\"utf-8\""))
	
	// Escape ampersands that are not part of an escape sequence (&...;).
	// In some chains they forgot to escape them and it annoys the XML parser.
	newText = regexp.MustCompile("&([^#a-z]|[a-z]+[^a-z;])").ReplaceAll(
			newText, []byte("&amp;$1"))
	
	// Quote unquoted attributes (Bitan has unquoted counts in their promo
	// files).
	newText = regexp.MustCompile("(\\w+=)(\\w+)").ReplaceAll(
			newText, []byte("$1\"$2\""))
	
	return newText, nil
}

