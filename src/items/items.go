// Parses price XMLs.
package main

import (
	"os"
	"fmt"
	"myflag"
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
	
	// Parse items.
	items, err := pricesParser.parse(data)
	pe(items, err)
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
	typ *string
	check *bool
	help bool
}

func parseArgs() error {
	args.file = myflag.String("file", "f", "path", "File to read from.", "")
	args.typ = myflag.String("type", "t", "type",
			"File type ('prices', 'stores' or 'promos').", "")
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
	
	return nil
}

var help =
`Parses XML files for the supermarket prices projects.

Usage:
items [-c] -t file-type -f file

Arguments:`




