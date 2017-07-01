// Parses price XMLs.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fluhus/gostuff/ezpprof"
	"github.com/fluhus/prices/bouncer"
	"github.com/fluhus/prices/serializer"
)

// Determines whether CPU profiling should be performed.
const profileCpu = false

// Determines whether memory profiling should be performed.
const profileMem = false

func main() {
	parseArgs()

	// Start profiling?
	if profileCpu {
		ezpprof.Start(filepath.Join(args.OutDir, "items.cpu.pprof"))
		defer ezpprof.Stop()
	}
	if profileMem {
		defer func() {
			runtime.GC()
			ezpprof.Heap(filepath.Join(args.OutDir, "items.mem.pprof"))
		}()
	}

	// Init bouncer.
	if !args.Check {
		bouncer.Initialize(args.OutDir)
		defer bouncer.Finalize()
	}

	// Prepare threads.
	numOfThreads := runtime.NumCPU()
	if numOfThreads > 16 {
		numOfThreads = 16
	}
	fmt.Println("Running on", numOfThreads, "threads.")
	fileChan := make(chan string, numOfThreads)
	doneChan := make(chan int, numOfThreads)
	errChan := make(chan error, numOfThreads)

	go func() { // Pushes file names for parsers.
		for _, file := range args.Files {
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

	// Start parser threads.
	for i := 0; i < numOfThreads; i++ {
		go func() {
			for file := range fileChan {
				err := processFile(file)
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

// Does the entire processing for a single file.
func processFile(file string) error {
	// Extract data-type, timestamp and chain-ID.
	typ := fileType(file)
	if typ == "" {
		return fmt.Errorf(
			"Could not infer data type (stores/prices/promos).")
	}

	// TODO(amit): Move timestamp to the data map.
	tim := fileTimestamp(file)
	if tim == -1 {
		return fmt.Errorf("Could not infer timestamp.")
	}

	// Attempt to read an already serialized file.
	var r reporter
	if !args.Check {
		r = reporters[typ]
	}

	var err error
	if !args.ForceRaw {
		err = reportSerializedFile(file+".items", r, tim)
	}

	// Parse raw file.
	if args.ForceRaw || err != nil {
		err = parseFile(file, parsers[typ])
		if err != nil {
			return err
		}
	}

	return reportSerializedFile(file+".items", r, tim)
}

// Parses a raw data file and serializes the result.
func parseFile(file string, prsr *parser) error {
	// Load input XML.
	data, err := load(file)
	if err != nil {
		return fmt.Errorf("Error reading raw file: %v", err)
	}

	// Make syntax & encoding corrections.
	data = correctXml(data)

	// Passing chain-ID because Co-Op don't include that field in their
	// XMLs.
	chainId := fileChainId(file)
	items, err := prsr.parse(data, map[string]string{"chain_id": chainId})
	if err != nil {
		return fmt.Errorf("Error parsing file: %v", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("Error parsing file: 0 items found.")
	}

	// Save processed file.
	meta := map[string]string{"version": parserVersion}
	items = append([]map[string]string{meta}, items...)

	err = serializer.Serialize(file+".items", items)
	if err != nil {
		return fmt.Errorf("Error serializing: %v", err)
	}

	return nil
}

// Reads a serialized file and reports it to the given reporter. If reporter
// is nil, reads without reporting (for check mode).
func reportSerializedFile(file string, r reporter, tim int64) error {
	d := serializer.NewDeserializer(file)

	// Read metadata.
	meta := d.Next()
	if meta == nil {
		return d.Err()
	}
	if meta["version"] != parserVersion {
		return fmt.Errorf("Mismatching parser version: expected '%s' actual '%s'.",
			parserVersion, meta["version"])
	}

	// Go over items.
	for item := d.Next(); item != nil; item = d.Next() {
		if r != nil {
			r([]map[string]string{item}, tim)
		}
	}

	// EOF is sababa, but other errors should be reported.
	if d.Err() == io.EOF {
		return nil
	} else {
		return d.Err()
	}
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
	match := regexp.MustCompile("(\\D|^)(20\\d{10})(\\D|$)").FindStringSubmatch(filepath.Base(file))
	if match == nil || len(match[2]) != 12 {
		return -1
	}
	digits := match[2]
	year, _ := strconv.ParseInt(digits[0:4], 10, 64)
	month, _ := strconv.ParseInt(digits[4:6], 10, 64)
	day, _ := strconv.ParseInt(digits[6:8], 10, 64)
	hour, _ := strconv.ParseInt(digits[8:10], 10, 64)
	minute, _ := strconv.ParseInt(digits[10:12], 10, 64)
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
