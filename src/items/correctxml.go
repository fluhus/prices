package main

// Functionality for correcting XML syntax and encoding errors.

import (
	"bytes"
	"golang.org/x/net/html/charset"
	"io/ioutil"
)

// Converts the given XML to utf-8, and corrects some syntax errors that the
// publishers make.
func correctXml(text []byte) ([]byte, error) {
	text = correctEncodingToUtf8(text)
	text = correctGibberish(text)
	text = correctUnquotedAttrs(text)
	text = correctEncodingField(text)
	text = correctAmpersands(text)

	return text, nil
}

// Encodes the given text in UTF-8.
func correctEncodingToUtf8(text []byte) []byte {
	r, _ := charset.NewReader(bytes.NewBuffer(text), "application/xml")
	text, _ = ioutil.ReadAll(r)
	return text
}

// Some Gibberish will not convert to UTF-8, so this function converts it
// manually.
func correctGibberish(text []byte) []byte {
	for i := 0; i < len(text)-1; i++ {
		if text[i] == 195 && text[i+1] >= 160 && text[i+1] <= 186 {
			text[i] = 215
			text[i+1] -= 16
		}
	}
	return text
}

// Replaces encoding attribute value with utf-8.
func correctEncodingField(text []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(text)))

	for i := 0; i < len(text); i++ {
		buf.WriteByte(text[i])

		// If found an encoding field.
		if text[i] == '=' && i >= 8 && string(text[i-8:i+1]) == "encoding=" {
			buf.WriteString("\"utf-8\"")

			// Advance to after end of field.
			i += 2
			for text[i] != '"' {
				i++
			}
		}
	}

	return buf.Bytes()
}

// Quote unquoted attributes (Bitan has unquoted counts in their promo
// files).
func correctUnquotedAttrs(text []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(text)))

	for i := 0; i < len(text); i++ {
		buf.WriteByte(text[i])

		if text[i] == '=' && i >= 1 && i < len(text)-1 &&
			isLetter(text[i-1]) && isAlphaNum(text[i+1]) {
			buf.WriteByte('"')
			for i < len(text)-1 && isAlphaNum(text[i+1]) {
				i++
				buf.WriteByte(text[i])
			}
			buf.WriteByte('"')
		}
	}

	return buf.Bytes()
}

// Escapes ampersands that are not part of an escape sequence (&...;).
// In some chains they forgot to escape them and it annoys the XML parser.
func correctAmpersands(text []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(text)))

	for i := 0; i < len(text); i++ {
		buf.WriteByte(text[i])

		if text[i] == '&' && i < len(text)-1 {
			suffix := text[i+1:]
			if suffix[0] == '#' {
				continue
			}
			for j := range suffix {
				if !isLetter(suffix[j]) {
					if suffix[j] != ';' {
						buf.WriteString("amp;")
					}
					break
				}
			}
		}
	}

	return buf.Bytes()
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphaNum(b byte) bool {
	return isLetter(b) || isDigit(b)
}
