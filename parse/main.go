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
	"sync"

	"github.com/fluhus/gostuff/ezpprof"
	"github.com/fluhus/prices/parse/bouncer"
	"github.com/fluhus/prices/serializer"
)

// TODO(amit): Switch to logging with the log package instead of prints.
// TODO(amit): Change ".items" suffix to something more self-descriptive. Perhaps ".parsed"?

const (
	profileCpu       = false    // Determines whether CPU profiling should be performed.
	profileMem       = false    // Determines whether memory profiling should be performed.
	parsedFileSuffix = ".items" // Suffix of parsed intermediates.
)

var (
	numOfThreads int // Shared across functions.
)

// What goes on here:
// 1. Handling some logistics of input files and threading.
// 2. Parsing the raw XMLs and writing parsed data to intermediate files.
// 3. Going over the intermediates and reporting them as table rows.
//
// Before that, it parsed and reported each file as a whole. That caused
// out-of-memory crashes, since parsing and reporting each takes a lot of
// memory. So doing them serially helps reducing the memory consumption of the
// process.
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

	// Prepare thread stuff.
	numOfThreads = runtime.NumCPU()
	errChan := make(chan error, numOfThreads)
	var wait sync.WaitGroup

	pe("Running on", numOfThreads, "threads.")

	// Prints errors to stderr. Each processed file is reported here, including success.
	ndone := 0 // TODO(amit): Find a better solution than this common counter.
	go func() {
		defer wait.Done()
		for err := range errChan {
			ndone++
			pef("%v (%v/%v %v%%)\n", err, ndone, len(inputFiles), ndone*100/len(inputFiles))
		}
	}()
	defer func() {
		wait.Add(1)
		close(errChan)
		wait.Wait()
	}()

	// Parse raw XMLs.
	pe("Parsing raw data.")
	fileChan := inputFilesChan(inputFiles)
	for i := 0; i < numOfThreads; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for file := range fileChan {
				err := parseFile(file.file)
				if err != nil {
					errChan <- fmt.Errorf("%v %s", err, file.file)
					continue
				}
				errChan <- fmt.Errorf("success parsing %s", file.file)
			}
		}()
	}
	wait.Wait()

	if args.Check {
		return
	}

	// Init bouncer.
	bouncer.Initialize(args.OutDir)
	defer bouncer.Finalize()

	// Report parsed data into tables.
	pe("Creating tables.")
	ndone = 0
	fileChan = inputFilesChan(inputFiles)
	for i := 0; i < numOfThreads; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for file := range fileChan {
				pfile := file.file + parsedFileSuffix // Name of parsed file.
				if !fileExists(pfile) {
					errChan <- fmt.Errorf("no parsed file for %s", file.file)
					continue
				}
				err := reportParsedFile(pfile, file.time)
				if err != nil {
					errChan <- fmt.Errorf("%v %s", err, file.file)
					continue
				}
				errChan <- fmt.Errorf("success reporting %s", file.file)
			}
		}()
	}
	wait.Wait()
}

// pe is Println to stderr.
func pe(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

// pef is Printf to stderr.
func pef(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
}

// parseFile parses a raw data file and serializes the result. Skips if a
// serialized output already exists and not force.
func parseFile(file string) error {
	// Check if already parsed.
	if !args.ForceRaw && fileExists(file+parsedFileSuffix) {
		// TODO(amit): Log?
		return nil
	}

	// Check file type.
	typ := fileType(file)
	if typ == "" {
		return fmt.Errorf("failed to infer data type (stores/prices/promos).")
	}

	// Load input XML.
	data, err := load(file)
	if err != nil {
		return fmt.Errorf("failed to read raw file: %v", err)
	}

	// Make syntax & encoding corrections.
	data = correctXml(data)

	// Passing chain-ID because Co-Op don't include that field in their
	// XMLs.
	chainId := fileChainId(file)
	items, err := parsers[typ].parse(data, map[string]string{"chain_id": chainId})
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("failed to parse file: 0 items found.")
	}

	// Save processed file.
	err = serializer.Serialize(file+parsedFileSuffix, items)
	if err != nil {
		return fmt.Errorf("failed to serialize: %v", err)
	}

	return nil
}

// reportParsedFile reads a serialized file and reports it.
func reportParsedFile(file string, tim int64) error {
	typ := fileType(file)
	if typ == "" {
		// Should not happen because parser wouldn't have parsed if no type.
		panic(fmt.Sprintf("No file type for serialized file. %v", file))
	}

	r := reporters[typ]
	d := serializer.NewDeserializer(file)

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

// fileType infers the type of data in the given file. Can be a full path.
// Returns either "prices", "stores", "promos", or an empty string if cannot
// infer.
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

// fileChainId infers the chain-ID of a file according to its name. Returns an
// empty string if failed.
func fileChainId(file string) string {
	match := regexp.MustCompile("\\D(7290\\d+)").FindStringSubmatch(file)
	if match == nil || len(match[1]) != 13 {
		return ""
	}

	return match[1]
}

// fileExists checks if a file or directory exists.
func fileExists(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}

// inputFilesChan returns a channel that gives the input files by their order,
// and closes it when they are over.
func inputFilesChan(files []*fileAndTime) chan *fileAndTime {
	result := make(chan *fileAndTime, numOfThreads*1000)
	go func() {
		for _, f := range files {
			result <- f
		}
		close(result)
	}()
	return result
}
