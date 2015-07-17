// Parses price XMLs.
package main

import (
	"os"
	"fmt"
	"myflag"
	"strings"
	"path/filepath"
)

func main() {
	// Handle arguments.
	err := parseArgs()
	if err != nil {
		pe("Error parsing arguments:", err)
		os.Exit(1)
	}
	if args.help {
		pe(help)
		pe(myflag.Help())
		os.Exit(1)
	}
	
	// Read input XML.
	data, err := load(*args.file)
	if err != nil {
		pe("Error reading input:", err)
		os.Exit(2)
	}
	
	// Convert to utf-8.
	data, err = toUtf8(data)
	if err != nil {
		pe("Error converting encoding:", err)
		os.Exit(2)
	}
	
	// Parse items.
	items, err := parsers[args.typ].parse(data)
	if err != nil {
		pe("Error parsing file:", err)
		os.Exit(2)
	}
	if len(items) == 0 {
		pe("Error parsing file: 0 items found.")
		os.Exit(2)
	}
	
	if !*args.check {
		fmt.Printf("%s", sqlers[args.typ].toSql(items))
	}
}

// Println to stderr.
func pe(a ...interface{}) {
	fmt.Println(a...)
}

// Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

var args struct {
	file *string
	typ string
	check *bool
	help bool
}

func parseArgs() error {
	args.file = myflag.String("file", "f", "path", "File to read from.", "")
	args.check = myflag.Bool("check", "c",
			"Only check file, do not print SQL statements.", false)
	
	err := myflag.Parse()
	if err != nil {
		return err
	}
	if !myflag.HasAny() {
		args.help = true
		return nil
	}
	if *args.file == "" {
		return fmt.Errorf("No input file supplied.")
	}
	
	// Infer data type.
	base := filepath.Base(*args.file)
	switch {
	case strings.HasPrefix(base, "Price"):
		args.typ = "prices"
	case strings.HasPrefix(base, "Store"):
		args.typ = "stores"
	case strings.HasPrefix(base, "Promo"):
		args.typ = "promos"
	default:
		return fmt.Errorf("Could not infer data type (stores, prices, promos).")
	}
	
	return nil
}

var help =
`Parses XML files for the supermarket prices projects.

Usage:
items [-c] -f file

Arguments:`




