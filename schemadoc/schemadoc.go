// Command schemadoc creates documentation from the DB schema.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	htemplate "html/template"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/fluhus/gostuff/flug"
)

func main() {
	parseArgs()

	// Read schema.
	pe("Reading SQLite schema from stdin...\n(run with -h argument for help)")
	textBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		pe("Error reading schema:", err)
		os.Exit(2)
	}
	if args.Verbose {
		pe("Read", len(textBytes), "characters.")
	}

	// Parse tables.
	text := string(textBytes)
	text = strings.Replace(text, "\r", "", -1)
	db, err := parseSchema(text)
	if err != nil {
		pe("Error parsing schema:", err)
		os.Exit(2)
	}
	if args.Verbose {
		pe("Parsed", len(db.Tables), "tables")
	}

	switch args.Format {
	case "text":
		fmt.Println(db)
	case "html":
		fmt.Println(db.html())
	case "latex":
		fmt.Println(db.latex())
	case "markdown":
		fmt.Println(db.markdown())
	case "json":
		j, err := json.MarshalIndent(db, "", "\t")
		if err != nil {
			pe("Failed to convert to JSON:", err)
			os.Exit(2)
		}
		fmt.Println(string(j))
	default:
		pe("NOT SUPPORTED YET: " + args.Format)
		os.Exit(2)
	}
}

var args = struct {
	Help    bool   `flug:"h,Show help message and exit."`
	Format  string `flug:"f,Output format: text, html, latex, json or markdown."`
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
	case "text", "html", "latex", "json", "markdown":
	default:
		pef("Error: unsupported format: %q\n", args.Format)
		os.Exit(1)
	}
}

// A schema is a completely parsed schema of the database.
type schema struct {
	Doc    string   `json:"doc"`
	Tables []*table `json:"tables"`
}

// A table is a single table in the database.
type table struct {
	Name   string   `json:"name"`
	Doc    string   `json:"doc"`
	Fields []*field `json:"fields"`
}

// A field is a single column in a table.
type field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Safe bool   `json:"safe"`
}

// parseSchema parses an SQLite schema and returns its object representation.
func parseSchema(text string) (*schema, error) {
	// Split to tables.
	re := regexp.MustCompile("(?ims)^CREATE TABLE (\\w*) \\(\\s*(.*?)\\s*^\\);$")
	matches := re.FindAllStringSubmatch(text, -1)
	if args.Verbose {
		pe("Found", len(matches), "tables.")
	}

	names := make([]string, len(matches))
	texts := make([]string, len(matches))
	for i := range matches {
		names[i] = matches[i][1]
		texts[i] = matches[i][2]
	}

	result := &schema{}
	var tables []*table
	for i := range texts {
		if args.Verbose {
			pe("Parsing table:", names[i])
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
func parseTable(text, name string) (*table, error) {
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
	rows := splitter.Split(text, -1)

	// Parse!
	for _, row := range rows {
		switch {
		case tableDoc.MatchString(row):
			match := tableDoc.FindStringSubmatch(row)
			if len(result.Doc) > 0 && len(match[1]) > 0 && result.Doc[len(result.Doc)-1] != '\n' {
				result.Doc += " "
			} else if len(result.Doc) > 0 && len(match[1]) == 0 {
				result.Doc += "\n\n"
			}
			result.Doc += match[1]

		case fieldDoc.MatchString(row):
			match := fieldDoc.FindStringSubmatch(row)

			// There must be a field to document.
			if len(result.Fields) == 0 {
				return nil, fmt.Errorf("field doc with no fields before it: %s",
					row)
			}

			f := result.Fields[len(result.Fields)-1]
			if len(f.Doc) > 0 && len(match[1]) > 0 {
				f.Doc += " "
			}
			f.Doc += match[1]
			if len(match[2]) > 0 {
				f.Safe = true
			}

		case fieldRow.MatchString(row):
			f := &field{}
			match := fieldRow.FindStringSubmatch(row)

			// Check if stumbled on a unique or a check.
			if len(match[1]) == 0 || nonFieldName.MatchString(match[1]) {
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
			return nil, fmt.Errorf("row doesn't match any pattern: %s", row)
		}
	}

	return result, nil
}

// String returns a string representation of a schema.
func (s *schema) String() string {
	tmp := template.Must(template.New("").Parse(textTemplate))
	buf := bytes.NewBuffer(nil)
	err := tmp.Execute(buf, s)
	if err != nil {
		return "Error: " + err.Error()
	}
	return buf.String()
}

// html returns an HTML representation of a schema.
func (s *schema) html() string {
	tmp := htemplate.Must(htemplate.New("").Parse(htmlTemplate))
	buf := bytes.NewBuffer(nil)
	err := tmp.Execute(buf, s)
	if err != nil {
		return "Error: " + err.Error()
	}
	return buf.String()
}

// latex returns a LaTeX representation of a schema.
func (s *schema) latex() string {
	tmp := template.Must(template.
		New("").
		Funcs(template.FuncMap{"latex": escapeLatex}).
		Parse(latexTemplate))
	buf := bytes.NewBuffer(nil)
	err := tmp.Execute(buf, s)
	if err != nil {
		return "Error: " + err.Error()
	}
	return buf.String()
}

// markdown returns a Markdown representation of a schema.
func (s *schema) markdown() string {
	tmp := template.Must(template.New("").Parse(markdownTemplate))
	buf := bytes.NewBuffer(nil)
	err := tmp.Execute(buf, s)
	if err != nil {
		return "Error: " + err.Error()
	}
	return buf.String()
}

// latexQuotes contains characters that should be escaped in LaTeX.
var latexQuotes = map[string]string{
	"&": "\\&",
	"_": "\\_",
	"%": "\\%",
	"{": "\\{",
	"}": "\\}",
}

// escapeLatex takes raw text and escapes special LaTeX characters.
func escapeLatex(text string) string {
	for s, r := range latexQuotes {
		text = strings.Replace(text, s, r, -1)
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

var textTemplate = `GENERAL INFORMATION
{{.Doc}}

{{range .Tables -}}
TABLE: {{.Name}}
{{.Doc}}

FIELDS:
{{- range .Fields}}
  {{.Name}}: {{.Doc}}{{if .Safe}} (safe){{end}}
{{- end}}

{{end}}`

var htmlTemplate = `<h3>General Information</h3>
<div>{{.Doc}}</div>
{{range .Tables -}}
<h3>{{.Name}}</h3>
{{.Doc}}
<div class="panel panel-default">
<table class="table table-bordered">
  <tr>
    <th>Field</th>
    <th>Safe</th>
    <th style="width:100%">Description</th>
  </tr>
  {{range .Fields -}}
  <tr>
    <td><strong>{{.Name}}</strong></td>
    <td><span {{if .Safe}}class="glyphicon glyphicon-ok"{{end}} aria-hidden="true"></span></td>
    <td>{{.Doc}}</td>
  </tr>
  {{- end}}
</table>
</div>
{{end}}`

var latexTemplate = `\\subsection{General Information}
{{latex .Doc}}

{{range .Tables -}}
\\subsection{ {{- latex .Name -}} }
{{latex .Doc}}

\\begin{tabularx}{\\linewidth}{|l|X|}
\\hline Field & Description \\\\
{{range .Fields -}}
\\hline {{latex .Name}} & {{latex .Doc}} \\\\
{{end -}}
\\hline \\end{tabularx}

{{end}}`

var markdownTemplate = `Database Schema
===============

General Information
-------------------

{{.Doc}}

Tables
------
{{range .Tables}}
### {{.Name}}

{{.Doc}}

**Fields**
{{range .Fields}}
* **{{.Name}}:** {{.Doc}}{{if .Safe}} (safe){{end}}
{{- end}}
{{- end}}`
