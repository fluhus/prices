// Parses price XMLs.
package main

import (
	"bouncer"
	"fmt"
	"github.com/fluhus/gostuff/ezpprof"
	"github.com/fluhus/gostuff/gobz"
	"io/ioutil"
	"myflag"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Determines whether CPU profiling should be performed.
const profileCpu = false

// Determines whether memory profiling should be performed.
const profileMem = false

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

	// Start profiling?
	if profileCpu {
		ezpprof.Start(filepath.Join(args.outDir, "items.cpu.pprof"))
		defer ezpprof.Stop()
	}
	if profileMem {
		defer func() {
			runtime.GC()
			ezpprof.Heap(filepath.Join(args.outDir, "items.mem.pprof"))
		}()
	}

	// Init bouncer.
	if !args.check {
		bouncer.Initialize(args.outDir)
		defer bouncer.Finalize()
	}

	// Prepare threads.
	numOfThreads := runtime.NumCPU()
	fmt.Println("Running on", numOfThreads, "threads.")
	runtime.GOMAXPROCS(numOfThreads)
	fileChan := make(chan string, numOfThreads)
	doneChan := make(chan int, numOfThreads)
	errChan := make(chan error, numOfThreads)

	go func() { // Pushes file names for parsers.
		for _, file := range args.files {
			fileChan <- file
		}
		close(fileChan)
	}()

	go func() { // Prints errors to stderr.
		for err := range errChan {
			pe(err)
		}
		doneChan <- 0
	}()

	// Go over files.
	for i := 0; i < numOfThreads; i++ {
		go func() {
			for file := range fileChan {
				err := parseFile(file)
				if err != nil {
					errChan <- fmt.Errorf("%v %s", err, file)
					continue
				} else {
					errChan <- fmt.Errorf("Success %s", file)
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
	close(errChan)
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
	files     []string
	check     bool
	outDir    string
	forceRaw  bool
	serialize bool
	help      bool
}

func parseArgs() error {
	// Set flags.
	check := myflag.Bool("check", "c",
		"Only check input files, do not create output tables.", false)
	filesFile := myflag.String("in", "i", "path",
		"A file that contains a list of input files, one per line.", "")
	outDir := myflag.String("out", "o", "path",
		"Output directory. Default is current.", ".")
	forceRaw := myflag.Bool("force-raw", "f",
		"Force parsing of raw files, instead of reading serialized data.",
		false)
	serialize := myflag.Bool("serialize", "s",
		"Create serialized files of parsed data, for faster loading in "+
			"the next run. Generated files will have the .gobz suffix.",
		false)

	// Parse flags.
	err := myflag.Parse()
	if err != nil {
		return err
	}
	if !myflag.HasAny() {
		args.help = true
		return nil
	}

	args.outDir = *outDir
	args.check = *check
	args.forceRaw = *forceRaw
	args.serialize = *serialize
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

// Does the entire processing for a single file.
func parseFile(file string) error {
	// Extract data-type, timestamp and chain-ID.
	typ := fileType(file)
	if typ == "" {
		return fmt.Errorf(
			"Could not infer data type (stores/prices/promos).")
	}

	tim := fileTimestamp(file)
	if tim == -1 {
		return fmt.Errorf("Could not infer timestamp.")
	}

	chainId := fileChainId(file)

	// Attempt to read an already serialized file.
	var items []map[string]string
	var err error
	if !args.forceRaw {
		err = gobz.Load(file+".gobz", &items)
	}

	// Parse raw file.
	if args.forceRaw || err != nil || len(items) == 0 ||
		items[0]["version"] != parserVersion {
		// Load input XML.
		data, err := load(file)
		if err != nil {
			return fmt.Errorf("Error reading input: %v", err)
		}

		// Make syntax & encoding corrections.
		data, err = correctXml(data)
		if err != nil {
			return fmt.Errorf("Error correcting XML: %v", err)
		}

		// Parse items.
		prsr := parsers[typ]
		if prsr == nil {
			panic("Nil parser for type '" + typ + "'.")
		}

		// Passing chain-ID because Co-Op don't include that field in their
		// XMLs.
		items, err = prsr.parse(data, map[string]string{"chain_id": chainId})
		if err != nil {
			return fmt.Errorf("Error parsing file: %v", err)
		}
		if len(items) <= 1 { // Only version item, no other data.
			return fmt.Errorf("Error parsing file: 0 items found.")
		}

		// Save processed file.
		if args.serialize {
			err = gobz.Save(file+".gobz", items)
			if err != nil {
				return fmt.Errorf("Error serializing: %v", err)
			}
		}
	}

	if args.check {
		return nil
	}

	reporters[typ](items[1:], tim) // Skip version item.
	return nil
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

var help = `Parses XML files for the supermarket price project.

Outputs TSV text files to the output directory. Supports XML, ZIP and GZ
formats. Do not use GOBZ files as input, use their prefix instead.

Usage:
items [OPTIONS] file1 file2 file3 ...
or
items [OPTIONS] -i <file with file-names>

Arguments:`
