// Command schemadoc creates documentation from the DB schema.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/fluhus/flug"
)

func main() {
	parseArgs()

	// Read schema.
	pe("Reading SQLite schema from stdin...\n(run with -help argument for help)")
	text, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading schema:", err)
		os.Exit(2)
	}

	// Parse tables.
	tables, err := parseSchema(text)
	if err != nil {
		pe("Error parsing schema:", err)
		os.Exit(2)
	}
	for _, t := range tables {
		switch args.Format {
		case "html":
			fmt.Printf("%s", t.html())
		case "latex":
			fmt.Printf("%s", t.latex())
		default:
			pe("NOT SUPPORTED YET: " + args.Format)
			os.Exit(2)
		}
	}
}

var args = struct {
	Help   bool   `flug:"help,Show help message."`
	Format string `flug:"format,Output format: text, html, latex, or json."`
}{false, "text"}

// parseArgs parses command line arguments. Exits on error.
func parseArgs() {
	flug.Register(&args)
	flag.Parse()

	if args.Help {
		pe("Creates documentation from the DB schema.")
		pe("Reads from stdin and prints to stdout.")
		pe()
		pe("Arguments:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	switch args.Format {
	case "text", "html", "latex", "json":
	default:
		pef("Error: unsupported format: %q\n", args.Format)
		os.Exit(1)
	}
}

// A table is a single table in the database.
type table struct {
	name   []byte
	doc    []byte
	fields []*field
}

// A field is a single column in a table.
type field struct {
	name []byte
	typ  []byte
	doc  []byte
	safe bool
}

// parseSchema parses an SQLite schema and returns its object representation.
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

// parseTable parses an SQLite table and returns its object representation.
func parseTable(text, name []byte) (*table, error) {
	// Regexps for capturing different type of information.
	tableDoc := regexp.MustCompile("^--\\s*(.*)$")
	fieldRow := regexp.MustCompile("^\\s*(\\w*)\\s*(\\w*).*?(?:--\\s*(.*?)\\s*(\\(safe\\))?)?$")
	fieldDoc := regexp.MustCompile("^\\s+--\\s*(.*?)\\s*(\\(safe\\))?$")
	splitter := regexp.MustCompile("\r?\n")
	nonFieldName := regexp.MustCompile("(?i:UNIQUE|CHECK)")

	// Initialize table.
	result := &table{}
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

			f := result.fields[len(result.fields)-1]
			if len(f.doc) > 0 && len(match[1]) > 0 {
				f.doc = append(f.doc, ' ')
			}
			f.doc = append(f.doc, match[1]...)
			if len(match[2]) > 0 {
				f.safe = true
			}

		case fieldRow.Match(brow):
			f := &field{}
			match := fieldRow.FindSubmatch(brow)

			// Check if stumbled on a unique or a check.
			if len(match[1]) == 0 || nonFieldName.Match(match[1]) {
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

// String returns a string representation of a table, for debugging.
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

// html returns an HTML representation of a table.
func (t *table) html() []byte {
	buf := bytes.NewBuffer(nil)

	// Create title and doc.
	fmt.Fprintf(buf, "<h3>%s</h3>\n", t.name)
	fmt.Fprintf(buf, "%s\n", t.doc)

	// Create table and header.
	fmt.Fprintf(buf, "<div class=\"panel panel-default\">\n")
	fmt.Fprintf(buf, "<table class=\"table table-bordered\">\n")
	fmt.Fprintf(buf, "<tr><th>Field</th><th>Safe</th><th "+
		"style=\"width:100%%\">Description</th></tr>\n")

	// Print fields.
	for _, f := range t.fields {
		class := ""
		if f.safe {
			class = "glyphicon glyphicon-ok"
		}
		fmt.Fprintf(buf, "<tr><td><strong>%s<strong></td><td><span "+
			"class=\"%s\" aria-hidden=\"true\""+
			"></span></td><td>%s</td></tr>\n", f.name, class, f.doc)
	}

	// Finish table.
	fmt.Fprintf(buf, "</table>\n</div>\n")

	return buf.Bytes()
}

// latex returns a LaTeX representation of a table.
func (t *table) latex() []byte {
	buf := bytes.NewBuffer(nil)

	// Create title and doc.
	fmt.Fprintf(buf, "\\subsection*{%s}\n", quoteLatex(t.name))
	fmt.Fprintf(buf, "%s\n", quoteLatex(t.doc))

	// Create table and header.
	fmt.Fprintf(buf, "\\begin{tabularx}{\\linewidth}{|l|X|}\n")
	fmt.Fprintf(buf, "\\hline Field & Description \\\\\n")

	// Print fields.
	for _, f := range t.fields {
		fmt.Fprintf(buf, "\\hline %s & %s \\\\\n",
			quoteLatex(f.name), quoteLatex(f.doc))
	}

	// Finish table.
	fmt.Fprintf(buf, "\\hline \\end{tabularx}\n\n")

	return buf.Bytes()
}

// latexQuotes contains characters that should be escaped in LaTeX.
var latexQuotes = map[string]string{
	"&": "\\&",
	"_": "\\_",
	"%": "\\%",
}

// quoteLatex takes raw text and escapes special LaTeX characters.
func quoteLatex(text []byte) []byte {
	for s, r := range latexQuotes {
		text = bytes.Replace(text, []byte(s), []byte(r), -1)
	}
	return text
}

// pe is a shorthand for Println to stderr.
func pe(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// pe is a shorthand for Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}
