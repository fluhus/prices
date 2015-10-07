package main

// Functionality for converting XML text encoding to UTF-8.

import (
	"bytes"
	"regexp"
	"io/ioutil"
	"golang.org/x/net/html/charset"
)

// Converts the given XML to utf-8.
func toUtf8(text []byte) ([]byte, error) {
	// Charset reader converts arbitrary text to UTF-8. Hurray!
	r, err := charset.NewReader(bytes.NewBuffer(text), "application/xml")
	if err != nil {
		return nil, err
	}
	
	newText, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	// Replace encoding field with utf-8.
	newText = regexp.MustCompile("encoding=\".*?\"").ReplaceAll(newText,
			[]byte("encoding=\"utf-8\""))
	
	// Escape ampersands that are not part of an escape sequence (&...;).
	// In some chains they forgot to escape them and it annoys the XML parser.
	newText = regexp.MustCompile("&([a-z]*[^a-z;])").ReplaceAll(
			newText, []byte("&amp;$1"))
	
	return newText, nil
}

