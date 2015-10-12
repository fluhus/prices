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
	"runtime"
	"io/ioutil"
	"mypprof"
)

// Determines whether CPU profiling should be performed.
const profile = false

func main() {
	// Start profiling?
	if profile {
		mypprof.Start("items.pprof")
		defer mypprof.Stop()
	}
	
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
	
	// Prepare threads.
	numOfThreads := runtime.NumCPU()
	runtime.GOMAXPROCS(numOfThreads)
	fileChan := make(chan string, numOfThreads)
	doneChan := make(chan int, numOfThreads)
	errChan := make(chan error, numOfThreads)
	sqlChan := make(chan []byte, numOfThreads)
	
	go func() {  // Pushes file names for parsers.
		for _, file := range args.files {
			fileChan <- file
		}
		close(fileChan)
	}()
	
	go func() {  // Prints generated SQL to stdout.
		for sql := range sqlChan {
			fmt.Printf("%s", sql)
		}
		doneChan <- 0
	}()
	
	go func() {  // Prints errors to stderr.
		for err := range errChan {
			pe(err)
		}
		doneChan <- 0
	}()
	
	// Parse files.
	for i := 0; i < numOfThreads; i++ {
		go func() {
			for file := range fileChan {
				// Turn file to SQL.
				sql, err := parseFile(file)
				if err != nil {
					errChan <- fmt.Errorf("%v %s", err, file)
					continue;
				} else {
					errChan <- fmt.Errorf("Success %s", file)
				}
				
				// Send SQL if not in check-mode.
				if !*args.check {
					sqlChan <- sql
				}
			}
			
			doneChan <- 0
		}()
	}
	
	// Wait for parser threads to finish.
	for i := 0; i < numOfThreads; i++ {
		<-doneChan
	}

	// Wait for printer threads.
	close(sqlChan)
	close(errChan)
	<-doneChan
	<-doneChan
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
	files []string
	check *bool
	help bool
}

func parseArgs() error {
	// Set flags.
	args.check = myflag.Bool("check", "c",
			"Only check file, do not print SQL statements.", false)
	filesFile := myflag.String("files", "f", "path",
			"A file that contains a list of files, one per line.", "")
	
	// Parse flags.
	err := myflag.Parse()
	if err != nil {
		return err
	}
	if !myflag.HasAny() {
		args.help = true
		return nil
	}
	
	args.files = myflag.Args()
	
	// Get file list from file.
	if *filesFile != "" {
		text, err := ioutil.ReadFile(*filesFile)
		if err != nil {
			return fmt.Errorf("Error reading file-list: %v", err)
		}
		
		if len(text) > 0 {
			args.files = strings.Split(string(text), "\n")
		}
	}
	
	if len(args.files) == 0 {
		return fmt.Errorf("No input files supplied.")
	}
	
	return nil
}

// Does the entire processing for a single file. Returns the SQL that should be
// passed to the database program.
func parseFile(file string) ([]byte, error) {
	// Extract data-type, timestamp and chain-ID.
	typ := fileType(file)
	if typ == "" {
		return nil, fmt.Errorf(
				"Could not infer data type (stores/prices/promos).")
	}
	
	tim := fileTimestamp(file)
	if tim == -1 {
		return nil, fmt.Errorf("Could not infer timestamp.")
	}
	
	chainId := fileChainId(file)
	
	// Read input XML.
	data, err := load(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading input: %v", err)
	}
	
	// Make syntax & encoding corrections.
	data, err = correctXml(data)
	if err != nil {
		return nil, fmt.Errorf("Error correcting XML: %v", err)
	}
	
	// Parse items.
	prsr := parsers[typ]
	if prsr == nil {
		panic("Nil parser for type '" + typ + "'.")
	}
	
	// Passing chain-ID because Co-Op don't include that field in their XMLs.
	items, err := prsr.parse(data, map[string]string{"chain_id":chainId})
	if err != nil {
		return nil, fmt.Errorf("Error parsing file: %v", err)
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

// Infers the chain-ID of a file according to its name. Returns an empty string
// if failed.
func fileChainId(file string) string {
	match := regexp.MustCompile("\\D(7290\\d+)").FindStringSubmatch(file)
	if match == nil || len(match[1]) != 13 {
		return ""
	}
	
	return match[1]
}

var help =
`Parses XML files for the supermarket prices projects.

Usage:
items [-c] file1 file2 file3 ...

Arguments:`




