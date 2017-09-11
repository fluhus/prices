// Parses price XMLs.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/fluhus/gostuff/ezpprof"
	"github.com/fluhus/prices/parse/bouncer"
	"github.com/fluhus/prices/serializer"
)

// TODO(amit): Switch to logging with the log package instead of prints.

// Determines whether CPU profiling should be performed.
const profileCpu = false

// Determines whether memory profiling should be performed.
const profileMem = false

func main() {
	parseArgs()
	inputFiles, err := organizeInputFiles()
	if err != nil {
		pe(err)
		os.Exit(2)
	}

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
	fmt.Println("Running on", numOfThreads, "threads.")
	fileChan := make(chan *fileAndTime, numOfThreads)
	doneChan := make(chan int, numOfThreads)
	errChan := make(chan error, numOfThreads)

	// Pushes file names to worker threads.
	go func() {
		for _, file := range inputFiles {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Prints errors to stderr. Each processed file is reported here, including success.
	go func() {
		i := 0
		for err := range errChan {
			i++
			pef("%v (%v/%v %v%%)\n", err, i, len(inputFiles), i*100/len(inputFiles))
		}
		doneChan <- 0
	}()

	// Start worker threads.
	for i := 0; i < numOfThreads; i++ {
		go func() {
			for file := range fileChan {
				err := processFile(file)
				if err != nil {
					errChan <- fmt.Errorf("%v %s", err, file.file)
					continue
				}
				errChan <- fmt.Errorf("success %s", file.file)
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
func processFile(file *fileAndTime) error {
	// Extract data-type.
	typ := fileType(file.file)
	if typ == "" {
		return fmt.Errorf(
			"could not infer data type (stores/prices/promos).")
	}

	// Attempt to read an already serialized file.
	var r reporter
	if !args.Check {
		r = reporters[typ]
	}

	// Try to load a parsed file.
	var err error
	if !args.ForceRaw {
		// TODO(amit): Change ".items" suffix to something more self-descriptive.
		// Perhaps ".parsed"?
		err = reportSerializedFile(file.file+".items", r, file.time)
	}

	// Parse raw file.
	if args.ForceRaw || err != nil {
		err = parseFile(file.file, parsers[typ])
		if err != nil {
			return err
		}
	}

	return reportSerializedFile(file.file+".items", r, file.time)
}

// Parses a raw data file and serializes the result.
func parseFile(file string, prsr *parser) error {
	// Load input XML.
	data, err := load(file)
	if err != nil {
		return fmt.Errorf("error reading raw file: %v", err)
	}

	// Make syntax & encoding corrections.
	data = correctXml(data)

	// Passing chain-ID because Co-Op don't include that field in their
	// XMLs.
	chainId := fileChainId(file)
	items, err := prsr.parse(data, map[string]string{"chain_id": chainId})
	if err != nil {
		return fmt.Errorf("error parsing file: %v", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("error parsing file: 0 items found.")
	}

	// Save processed file.
	meta := map[string]string{"version": parserVersion}
	items = append([]map[string]string{meta}, items...)

	err = serializer.Serialize(file+".items", items)
	if err != nil {
		return fmt.Errorf("error serializing: %v", err)
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
		// This means that the parser was updated after the file was serialized,
		// so this serialized file is invalid.
		return fmt.Errorf("mismatching parser version: expected '%s' actual '%s'.",
			parserVersion, meta["version"])
	}

	// Go over items.
	for item := d.Next(); item != nil; item = d.Next() {
		if r != nil {
			r([]map[string]string{item}, tim)
		}
	}

	// EOF is ok, but other errors should be reported.
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

// Infers the chain-ID of a file according to its name. Returns an empty string
// if failed.
func fileChainId(file string) string {
	match := regexp.MustCompile("\\D(7290\\d+)").FindStringSubmatch(file)
	if match == nil || len(match[1]) != 13 {
		return ""
	}

	return match[1]
}
