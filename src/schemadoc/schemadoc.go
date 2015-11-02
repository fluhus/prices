// Creates documentation from the DB schema.
package main

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"bytes"
)

func main() {
	// Read schema.
	pe("Reading from stdin...")
	text, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading schema:", err)
		os.Exit(1)
	}
	
	// Parse tables.
	pe("Parsing...")
	tables, err := parseSchema(text)
	if err != nil {
		pe("error:", err)
	} else {
		pe(tables[5])
	}
}

type table struct {
	name []byte
	doc []byte
	fields []*field
}

type field struct {
	name []byte
	typ []byte
	doc []byte
	safe bool
}

func parseSchema(text []byte) ([]*table, error) {
	// Split to tables.
	re := regexp.MustCompile("(?ims)^CREATE TABLE (\\w*) \\(\\s*(.*?)\\s*^\\);$")
	matches := re.FindAllSubmatch(text, -1)
	names := make([][]byte, len(matches))
	texts := make([][]byte, len(matches))
	for i := range matches {
		names[i] = matches[i][1]
		texts[i] = matches[i][2]
	}
	
	result := make([]*table, len(texts))
	for i := range texts {
		var err error
		result[i], err = parseTable(texts[i], names[i])
		if err != nil {
			return nil, fmt.Errorf("Table 1: %v", err)
		}
	}
	
	return result, nil
}

func parseTable(text, name []byte) (*table, error) {
	// Regexps for capturing different type of information.
	tableDoc := regexp.MustCompile("^--\\s*(.*)$")
	fieldRow := regexp.MustCompile("^\\s*(\\w*)\\s*(\\w*).*?(?:--\\s*(.*?)\\s*(\\(safe\\))?)?$")
	fieldDoc := regexp.MustCompile("^\\s+--\\s*(.*?)\\s*(\\(safe\\))?$")
	splitter := regexp.MustCompile("\r?\n")
	nonFieldName := regexp.MustCompile("(?i:UNIQUE|CHECK)")
	
	// Initialize table.
	result := &table {}
	result.name = name
	
	// Split to rows.
	rows := splitter.Split(string(text), -1)
	
	// Parse!
	for i := range rows {
		brow := []byte(rows[i])
		switch {
		case tableDoc.Match(brow):
			match := tableDoc.FindSubmatch(brow)
			if len(result.doc) > 0 && len(match[1]) > 0 {
				result.doc = append(result.doc, ' ')
			}
			result.doc = append(result.doc, match[1]...)
		
		case fieldDoc.Match(brow):
			match := fieldDoc.FindSubmatch(brow)
			
			// There must be a field to document.
			if len(result.fields) == 0 {
				return nil, fmt.Errorf("Field doc with no fields before it: %s",
						brow)
			}
			
			f := result.fields[len(result.fields) - 1]
			if len(f.doc) > 0 && len(match[1]) > 0 {
				f.doc = append(f.doc, ' ')
			}
			f.doc = append(f.doc, match[1]...)
			if len(match[2]) > 0 {
				f.safe = true
			}
		
		case fieldRow.Match(brow):
			f := &field {}
			match := fieldRow.FindSubmatch(brow)
			
			// Check if stumbled on a unique or a check.
			if nonFieldName.Match(match[1]) {
				continue
			}
			
			f.name = match[1]
			f.typ = match[2]
			f.doc = match[3]
			if len(match[4]) > 0 {
				f.safe = true
			}
			result.fields = append(result.fields, f)
		
		default:
			return nil, fmt.Errorf("Row doesn't match any pattern: %s", brow)
		}
	}
	
	return result, nil
}

func (t *table) String() string {
	buf := bytes.NewBuffer(nil)
	
	fmt.Fprintf(buf, "TABLE: %s\n", t.name)
	fmt.Fprintf(buf, "%s\n\n", t.doc)
	fmt.Fprintf(buf, "FIELDS:\n")
	if len(t.fields) == 0 {
		fmt.Fprintf(buf, "(none)\n")
	}
	for _, f := range t.fields {
		if f.safe {
			fmt.Fprintf(buf, "%s (%s, safe)\n", f.name, f.typ)
		} else {
			fmt.Fprintf(buf, "%s (%s)\n", f.name, f.typ)
		}
		if len(f.doc) > 0 {
			fmt.Fprintf(buf, "%s\n", f.doc)
		}
	}
	
	return buf.String()
}

func pe(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

