// Scrapes raw data from the different chains.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"flag"
	"path/filepath"
	"time"

	"github.com/fluhus/flug"
	"github.com/fluhus/prices/scrape/scrapers"
)

func main() {
	// Parse arguments.
	err := parseArgs()
	if err == noArgs {
		fmt.Fprintln(os.Stderr, help)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\n" + credit)
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

	for _, task := range tasks {
		// A task may be nil, to make a placeholder for a future scraper.
		if task == nil {
			continue
		}

		tt := time.Now()
		log.SetPrefix(task.name + " ")
		log.Printf("Starting %s.", task.name)

		err := task.scrp.Scrape(filepath.Join(args.Dir, "{{date}}", task.dir))
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

// A single scraping task.
type scrapingTask struct {
	scrp scrapers.Scraper // Performer of scraping.
	name string           // For logging.
	dir  string           // Directory to which to download files.
}

// Holds tasks to perform by the main program. Tasks will be performed
// sequentially. Use a nil value to make a placeholder, for chain counting.
var tasks = []*scrapingTask{
	&scrapingTask{scrapers.Cerberus("TivTaam", ""), "TivTaam", "tivtaam"},
	&scrapingTask{scrapers.Shufersal(), "Shufersal", "shufersal"},
	&scrapingTask{scrapers.Cerberus("DorAlon", ""), "DorAlon", "doralon"},
	&scrapingTask{scrapers.Cerberus("osherad", ""), "OsherAd", "osherad"},
	&scrapingTask{scrapers.Mega(), "Mega", "mega"},
	&scrapingTask{scrapers.Cerberus("HaziHinam", ""), "HaziHinam", "hazihinam"},
	&scrapingTask{scrapers.Cerberus("Keshet", ""), "Keshet", "keshet"},
	&scrapingTask{scrapers.Cerberus("RamiLevi", ""), "RamiLevi", "ramilevi"},
	&scrapingTask{scrapers.Cerberus("SuperDosh", ""), "SuperDosh", "superdosh"},
	&scrapingTask{scrapers.Cerberus("Yohananof", ""), "Yohananof", "yohananof"},
	&scrapingTask{scrapers.Eden(), "Eden", "eden"},
	&scrapingTask{scrapers.Bitan(), "Bitan", "bitan"},
	&scrapingTask{scrapers.Nibit(scrapers.Victory, 7), "Victory", "victory"},
	&scrapingTask{scrapers.Nibit(scrapers.Hashook, 7), "Hashook", "hashook"},
	&scrapingTask{scrapers.Nibit(scrapers.Lahav, 7), "Lahav", "lahav"},
	&scrapingTask{scrapers.Coop(), "Coop", "coop"},
	&scrapingTask{scrapers.Cerberus("freshmarket_sn", "f_efrd"), "Freshmarket", "freshmarket"},
	&scrapingTask{scrapers.Zolbegadol(), "ZolBegadol", "zolbegadol"},
}

// Returns the name that should be given to the log file.
func logFileName() string {
	t := time.Now()
	return fmt.Sprintf("Log-%d%02d%02d%02d%02d.txt",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

// Holds parsed command-line arguments.
var args struct {
	Dir    string  // Where to download files.
	Stdout bool   `flug:"stdout,Log to stdout instead of log file."`
	From   string `flug:"from,Download files from this time and on. Format: YYYYMMDDhhmm. (default download all files)"`
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

	// No args.
	if len(flag.Args()) == 0 {
		return noArgs
	}

	// Too many args.
	if len(flag.Args()) > 1 {
		return fmt.Errorf("To many arguments were given.")
	}

	args.Dir = flag.Args()[0]

	return nil
}

// Help message to display when run with no arguments.
var help = `Downloads price data from stores.

Usage:
prices <out dir>

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
	chains := ""
	for i := range tasks {
		if tasks[i] == nil {
			continue
		}
		if i > 0 {
			chains += ", "
		}
		chains += tasks[i].name
	}
	log.Print("Chains in this run: ", chains)
}
