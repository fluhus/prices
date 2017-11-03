// Command schemadoc creates documentation from the DB schema.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/fluhus/flug"
)

// TODO(amit): Make error messages start with lowercase.
// TODO(amit): Use templates in html and latex generation.
// TODO(amit): Add markdown?

func main() {
	parseArgs()

	// Read schema.
	pe("Reading SQLite schema from stdin...\n(run with -h argument for help)")
	text, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading schema:", err)
		os.Exit(2)
	}
	if args.Verbose {
		pe("Read", len(text), "characters.")
	}

	// Parse tables.
	text = bytes.Replace(text, []byte("\r"), nil, -1)
	db, err := parseSchema(text)
	if err != nil {
		pe("Error parsing schema:", err)
		os.Exit(2)
	}
	if args.Verbose {
		pe("Parsed", len(db.Tables), "tables")
	}

	for _, t := range db.Tables {
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
	Help    bool   `flug:"h,Show help message and exit."`
	Format  string `flug:"f,Output format: html, latex, or json."`
	Verbose bool   `flug:"v,Verbose, print debug messages."`
}{false, "text", false}

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
	case "html", "latex", "json":
	default:
		pef("Error: unsupported format: %q\n", args.Format)
		os.Exit(1)
	}
}

// A schema is a completely parsed schema of the database.
type schema struct {
	Doc    []byte
	Tables []*table
}

// A table is a single table in the database.
type table struct {
	Name   []byte
	Doc    []byte
	Fields []*field
}

// A field is a single column in a table.
type field struct {
	Name []byte
	Type []byte
	Doc  []byte
	Safe bool
}

// parseSchema parses an SQLite schema and returns its object representation.
func parseSchema(text []byte) (*schema, error) {
	// Split to tables.
	re := regexp.MustCompile("(?ims)^CREATE TABLE (\\w*) \\(\\s*(.*?)\\s*^\\);$")
	matches := re.FindAllSubmatch(text, -1)
	if args.Verbose {
		pe("Found", len(matches), "tables.")
	}

	names := make([][]byte, len(matches))
	texts := make([][]byte, len(matches))
	for i := range matches {
		names[i] = matches[i][1]
		texts[i] = matches[i][2]
	}

	result := &schema{}
	var tables []*table
	for i := range texts {
		if args.Verbose {
			pe("Parsing table:", string(names[i]))
		}

		t, err := parseTable(texts[i], names[i])
		if err != nil {
			return nil, fmt.Errorf("table %v: %v", i+1, err)
		}
		if string(t.Name) == "documentation" {
			result.Doc = t.Doc
			continue
		}
		tables = append(tables, t)
	}
	result.Tables = tables

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
	result.Name = name

	// Split to rows.
	rows := splitter.Split(string(text), -1)

	// Parse!
	for i := range rows {
		brow := []byte(rows[i])
		switch {
		case tableDoc.Match(brow):
			match := tableDoc.FindSubmatch(brow)
			if len(result.Doc) > 0 && len(match[1]) > 0 {
				result.Doc = append(result.Doc, ' ')
			}
			result.Doc = append(result.Doc, match[1]...)

		case fieldDoc.Match(brow):
			match := fieldDoc.FindSubmatch(brow)

			// There must be a field to document.
			if len(result.Fields) == 0 {
				return nil, fmt.Errorf("field doc with no fields before it: %s",
					brow)
			}

			f := result.Fields[len(result.Fields)-1]
			if len(f.Doc) > 0 && len(match[1]) > 0 {
				f.Doc = append(f.Doc, ' ')
			}
			f.Doc = append(f.Doc, match[1]...)
			if len(match[2]) > 0 {
				f.Safe = true
			}

		case fieldRow.Match(brow):
			f := &field{}
			match := fieldRow.FindSubmatch(brow)

			// Check if stumbled on a unique or a check.
			if len(match[1]) == 0 || nonFieldName.Match(match[1]) {
				continue
			}

			f.Name = match[1]
			f.Type = match[2]
			f.Doc = match[3]
			if len(match[4]) > 0 {
				f.Safe = true
			}
			result.Fields = append(result.Fields, f)

		default:
			return nil, fmt.Errorf("row doesn't match any pattern: %s", brow)
		}
	}

	return result, nil
}

// String returns a string representation of a table, for debugging.
func (s *schema) String() string {
	j, _ := json.Marshal(s)
	return string(j)
}

// html returns an HTML representation of a table.
func (t *table) html() []byte {
	buf := bytes.NewBuffer(nil)

	// Create title and doc.
	fmt.Fprintf(buf, "<h3>%s</h3>\n", t.Name)
	fmt.Fprintf(buf, "%s\n", t.Doc)

	// Create table and header.
	fmt.Fprintf(buf, "<div class=\"panel panel-default\">\n")
	fmt.Fprintf(buf, "<table class=\"table table-bordered\">\n")
	fmt.Fprintf(buf, "<tr><th>Field</th><th>Safe</th><th "+
		"style=\"width:100%%\">Description</th></tr>\n")

	// Print fields.
	for _, f := range t.Fields {
		class := ""
		if f.Safe {
			class = "glyphicon glyphicon-ok"
		}
		fmt.Fprintf(buf, "<tr><td><strong>%s<strong></td><td><span "+
			"class=\"%s\" aria-hidden=\"true\""+
			"></span></td><td>%s</td></tr>\n", f.Name, class, f.Doc)
	}

	// Finish table.
	fmt.Fprintf(buf, "</table>\n</div>\n")

	return buf.Bytes()
}

// latex returns a LaTeX representation of a table.
func (t *table) latex() []byte {
	buf := bytes.NewBuffer(nil)

	// Create title and doc.
	fmt.Fprintf(buf, "\\subsection*{%s}\n", quoteLatex(t.Name))
	fmt.Fprintf(buf, "%s\n", quoteLatex(t.Doc))

	// Create table and header.
	fmt.Fprintf(buf, "\\begin{tabularx}{\\linewidth}{|l|X|}\n")
	fmt.Fprintf(buf, "\\hline Field & Description \\\\\n")

	// Print fields.
	for _, f := range t.Fields {
		fmt.Fprintf(buf, "\\hline %s & %s \\\\\n",
			quoteLatex(f.Name), quoteLatex(f.Doc))
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
