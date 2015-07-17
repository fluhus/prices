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
	
	// Read input XML.
	data, err := load(*args.file)
	if err != nil {
		pe("Error reading input:", err, *args.file)
		os.Exit(2)
	}
	
	// Convert to utf-8.
	data, err = toUtf8(data)
	if err != nil {
		pe("Error converting encoding:", err, *args.file)
		os.Exit(2)
	}
	
	// Parse items.
	items, err := parsers[args.typ].parse(data)
	if err != nil {
		pe("Error parsing file:", err, *args.file)
		os.Exit(2)
	}
	if len(items) == 0 {
		pe("Error parsing file: 0 items found.", *args.file)
		os.Exit(2)
	}
	
	if !*args.check {
		fmt.Printf("%s", sqlers[args.typ].toSql(items, args.time))
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
	typ string
	time int64
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
		return fmt.Errorf("Could not infer data type (stores/prices/promos)." +
				" %s", *args.file)
	}
	
	// Infer timestamp.
	match := regexp.MustCompile("\\D(\\d+)\\D*$").FindStringSubmatch(
			*args.file)
	if match == nil || len(match[1]) != 12 {
		return fmt.Errorf("Could not infer timestamp. %s", *args.file)
	}
	year, _ := strconv.ParseInt(match[1][0:4], 10, 64)
	month, _ := strconv.ParseInt(match[1][4:6], 10, 64)
	day, _ := strconv.ParseInt(match[1][6:8], 10, 64)
	hour, _ := strconv.ParseInt(match[1][8:10], 10, 64)
	minute, _ := strconv.ParseInt(match[1][10:12], 10, 64)
	t := time.Date(int(year), time.Month(month), int(day), int(hour),
			int(minute), 0, 0, time.UTC)
	args.time = t.Unix()
	
	return nil
}

var help =
`Parses XML files for the supermarket prices projects.

Usage:
items [-c] -f file

Arguments:`




