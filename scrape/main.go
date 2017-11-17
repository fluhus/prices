// Scrapes raw data from the different chains.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fluhus/gostuff/flug"
	"github.com/fluhus/prices/scrape/scrapers"
)

func main() {
	// Parse arguments.
	err := parseArgs()
	if err == noArgs {
		fmt.Fprintln(os.Stderr, help)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n"+credit)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal("Failed to parse arguments: ", err)
	}

	// Open logging output file.
	if !args.Stdout {
		logsDir := filepath.Join(args.Dir, "logs")
		err = os.MkdirAll(logsDir, 0700)
		if err != nil {
			log.Fatal("Filed to create output dir:", err)
		}
		out, err := os.Create(filepath.Join(logsDir, logFileName()))
		if err != nil {
			log.Fatal("Error creating log file:", err)
		}
		defer out.Close()
		buf := bufio.NewWriter(out)
		defer buf.Flush()
		log.SetOutput(buf)
	}

	logWelcome()

	// Check that number of chains matches number of tasks.
	chainCount, err := scrapers.CountChains()
	if err != nil {
		log.Printf("Chain count error: %v", err)
	} else {
		if chainCount != len(tasks) {
			// TODO(amit): Improve this error message.
			log.Printf("Chain count error: Found %d chains but there are %d"+
				" tasks. To silence this error, place a nil placeholder"+
				" in the task list.", chainCount, len(tasks))
		}
	}

	// Perform scraping tasks.
	t := time.Now()

	for _, chain := range args.ChainList {
		scrp := tasks[chain]

		// A task may be nil, to make a placeholder for a future scraper.
		if scrp == nil {
			continue
		}

		tt := time.Now()
		log.SetPrefix(chain + " ")
		log.Printf("Starting %s.", chain)

		err := scrp.Scrape(filepath.Join(args.Dir, "{{date}}", chain))
		if err != nil {
			log.Printf("Finished with error: %v", err)
		} else {
			log.Printf("Finished successfully.")
		}

		log.Println("Time took:", time.Now().Sub(tt))
		log.SetPrefix("")
	}

	log.Printf("Operation is complete. Time took: %v", time.Now().Sub(t))
}

// Holds tasks to perform by the main program. Tasks will be performed ordered
// by flag value, or alphabetically if chains flag is empty.
//
// Use a nil value to make a placeholder, for chain counting.
var tasks = map[string]scrapers.Scraper{
	"tivtaam":     scrapers.Cerberus("TivTaam", ""),
	"shufersal":   scrapers.Shufersal(),
	"doralon":     scrapers.Cerberus("DorAlon", ""),
	"osherad":     scrapers.Cerberus("osherad", ""),
	"mega":        scrapers.Mega(),
	"hazihinam":   scrapers.Cerberus("HaziHinam", ""),
	"keshet":      scrapers.Cerberus("Keshet", ""),
	"ramilevi":    scrapers.Cerberus("RamiLevi", ""),
	"superdosh":   scrapers.Cerberus("SuperDosh", ""),
	"yohananof":   scrapers.Cerberus("Yohananof", ""),
	"eden":        scrapers.Eden(),
	"bitan":       scrapers.Bitan(),
	"victory":     scrapers.Nibit(scrapers.Victory, 7),
	"hashook":     scrapers.Nibit(scrapers.Hashook, 7),
	"lahav":       scrapers.Nibit(scrapers.Lahav, 7),
	"coop":        scrapers.Coop(),
	"freshmarket": scrapers.Cerberus("freshmarket_sn", "f_efrd"),
	"zolbegadol":  scrapers.Zolbegadol(),
}

// Returns the name that should be given to the log file.
func logFileName() string {
	t := time.Now()
	return fmt.Sprintf("Log-%d%02d%02d%02d%02d.txt",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

// Holds parsed command-line arguments.
var args struct {
	Dir       string   // Where to download files.
	ChainList []string // List of chain names to include in this run, parsed from Chains.
	Stdout    bool     `flug:"stdout,Log to stdout instead of log file."`
	From      string   `flug:"from,Download files from this time and on. Format: YYYYMMDDhhmm. (default download all files)"`
	Chains    string   `flug:"chains,Comma separated chain names to include in this run. (default all)"`
}

// Signifies that no args were given.
var noArgs = fmt.Errorf("")

// Parses arguments and places their values in the args struct. If an error
// returns, args are invalid.
func parseArgs() error {
	flug.Register(&args)
	flag.Parse()

	// Parse timestamp.
	if args.From != "" {
		err := scrapers.SetFromTimestamp(args.From)
		if err != nil {
			return err
		}
	}

	// Parse chains.
	if args.Chains == "" {
		for chain := range tasks {
			args.ChainList = append(args.ChainList, chain)
		}
		sort.Strings(args.ChainList)
	} else {
		// TODO(amit): Handle duplicates.
		args.ChainList = strings.Split(args.Chains, ",")
		for _, chain := range args.ChainList {
			if tasks[chain] == nil {
				return fmt.Errorf("unrecognized chain name: %q", chain)
			}
		}
	}

	// No args.
	if len(flag.Args()) == 0 {
		return noArgs
	}

	// Too many args.
	if len(flag.Args()) > 1 {
		return fmt.Errorf("too many arguments were given.")
	}

	args.Dir = flag.Args()[0]

	return nil
}

// Help message to display when run with no arguments.
var help = `Downloads price data from stores.

Usage:
scrape <out dir>

Flags:`

var credit = `Credit:
Based on the 'prices' project by Amit Lavon.
https://github.com/fluhus/prices`

// Prints a welcome message and usage instructions to the log.
func logWelcome() {
	log.Print("We have lift off!")

	// Print grep help.
	log.Print("To search for a specific chain use grep '^ChainName'.")
	log.Print("To search for errors, use grep 'error' (use tail to omit this line).")
	log.Print("To search for times, use grep 'took'.")

	// Print chain names.
	log.Print("Chains in this run: ", strings.Join(args.ChainList, ", "))
}
