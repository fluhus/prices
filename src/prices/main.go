package main

import (
	"aggregators"
	"time"
	"runtime"
	"log"
	"fmt"
	"bufio"
	"os"
	"path/filepath"
)

func main() {
	// --- Initialization stuff. ---
	// Set number of CPUs to max.
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Parse arguments.
	err := parseArgs()
	if err == noArgs {
		fmt.Print(help)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal("Failed to parse arguments: ", err)
	}
	
	// Open logging output file.
	logsDir := filepath.Join(args.dir, "logs")
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
	
	// --- Perform aggregation tasks. ---
	logWelcome()
	t := time.Now()
	
	for _, task := range tasks {
		tt := time.Now()
		log.SetPrefix(task.name + " ")
		
		err := task.agg.Aggregate(filepath.Join(args.dir, task.dir))
		if err != nil {
			log.Printf("Finished with error: %v", err)
		} else {
			log.Printf("Finished successfully.")
		}
		
		log.Println("Took", time.Now().Sub(tt))
		log.SetPrefix("")
	}
	
	log.Printf("Operation is complete. Took %v.", time.Now().Sub(t))
}

// A single aggregation task.
type aggTask struct {
	agg  aggregators.Aggregator  // Performer of aggregation.
	name string                  // For logging.
	dir  string                  // Directory to which to download files.
}

// Holds tasks to perform by the main program. Tasks will be performed
// sequentially.
var tasks = []*aggTask {
	// &aggTask{ aggregators.NewCerberusAggregator("doralon"),
			// "DorAlon", "doralon" },
	&aggTask{ aggregators.Cerberus("Keshet"), "Keshet", "keshet" },
	&aggTask{ aggregators.Shufersal(), "Shufersal", "shufersal" },
}

// Returns the name that should be given to the log file.
func logFileName() string {
	t := time.Now()
	return fmt.Sprintf("Log-%d%02d%02d%02d%02d.txt",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

// Holds parsed command-line arguments.
var args struct {
	dir string
}

// Signifies that no args were given.
var noArgs = fmt.Errorf("")

// Parses arguments and places their values in the args struct. If an error
// returns, args are invalid.
func parseArgs() error {
	a := os.Args[1:]  // Omit program name.
	
	// No args.
	if len(a) == 0 {
		return noArgs
	}
	
	// To many args.
	if len(a) > 1 {
		return fmt.Errorf("To many arguments were given.")
	}
	
	args.dir = a[0]
	
	return nil
}

// Help message to display when run with no arguments.
var help = `Downloads price data from stores.

Usage:
prices <out dir>
`

// Prints a welcome message and usage instructions to the log.
func logWelcome() {
	log.Print("We have lift off!")
	
	// Print chain names.
	chains := ""
	for i := range tasks {
		if i > 0 { chains += ", " }
		chains += tasks[i].name
	}
	
	log.Print("Chains in this run: ", chains)
	
	// Print grep help.
	log.Print("Each log message is prefixed by its chain name. To search for " +
			"messages from a specific chain use grep '^chain'. For " +
			"example: grep '^DorAlon'")
	log.Print("To search for errors, use grep 'error'.")
}

