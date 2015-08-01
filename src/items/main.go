// Parses price XMLs.
package main

import (
	"os"
	"fmt"
	"myflag"
	"strings"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
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
	
	sql, err := parseFile(*args.file)
	if err != nil {
		pe(err)
		os.Exit(2)
	}
	
	if !*args.check {
		fmt.Printf("%s", sql)
	}
}

// Println to stderr.
func pe(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

var args struct {
	file *string
	check *bool
	help bool
}

func parseArgs() error {
	// Set flags.
	args.file = myflag.String("file", "f", "path", "File to read from.", "")
	args.check = myflag.Bool("check", "c",
			"Only check file, do not print SQL statements.", false)
	
	// Parse flags.
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

func parseFile(file string) ([]byte, error) {
	// Extract data-type and timestamp.
	typ := fileType(file)
	if typ == "" {
		return nil,
				fmt.Errorf("Could not infer data type (stores/prices/promos).")
	}
	
	tim := fileTimestamp(file)
	if tim == -1 {
		return nil, fmt.Errorf("Could not infer timestamp.")
	}
	
	// Read input XML.
	data, err := load(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading input:", err)
	}
	
	// Convert to utf-8.
	data, err = toUtf8(data)
	if err != nil {
		return nil, fmt.Errorf("Error converting encoding:", err)
	}
	
	// Parse items.
	prsr := parsers[typ]
	if prsr == nil {
		panic("Nil parser for type '" + typ + "'.")
	}
	items, err := prsr.parse(data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing file:", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("Error parsing file: 0 items found.")
	}
	
	return sqlers[typ](items, tim), nil
}

// Infers the type of data in the given file. Can be a full path. Returns either
// "prices", "stores", "promos", or an empty string if cannot infer.
func fileType(file string) string {
	base := filepath.Base(file)
	switch {
	case strings.HasPrefix(base, "Price"):
		return "prices"
	case strings.HasPrefix(base, "Store"):
		return "stores"
	case strings.HasPrefix(base, "Promo"):
		return "promos"
	default:
		return ""
	}
}

// Infers the timestamp of a file according to its name. Returns -1 if failed.
func fileTimestamp(file string) int64 {
	match := regexp.MustCompile("\\D(201\\d+)").FindStringSubmatch(file)
	if match == nil || len(match[1]) != 12 {
		return -1
	}
	year, _ := strconv.ParseInt(match[1][0:4], 10, 64)
	month, _ := strconv.ParseInt(match[1][4:6], 10, 64)
	day, _ := strconv.ParseInt(match[1][6:8], 10, 64)
	hour, _ := strconv.ParseInt(match[1][8:10], 10, 64)
	minute, _ := strconv.ParseInt(match[1][10:12], 10, 64)
	t := time.Date(int(year), time.Month(month), int(day), int(hour),
			int(minute), 0, 0, time.UTC)
	
	return t.Unix()
}

var help =
`Parses XML files for the supermarket prices projects.

Usage:
items [-c] -f file

Arguments:`




