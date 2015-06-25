package main

import (
	"aggregators"
	"time"
	"runtime"
	"log"
	"fmt"
	"bufio"
	"myflag"
	"os"
	"path/filepath"
)

func main() {
	// Set number of CPUs to max.
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	// Parse arguments.
	err := parseArgs()
	if err == noArgs {
		fmt.Print(help)
		fmt.Print(myflag.Help())
		os.Exit(1)
	}
	if err != nil {
		log.Fatal("Failed to parse arguments: ", err)
	}
	
	// Open logging output file.
	if !*args.stdout {
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
	}
	
	logWelcome()
	
	// Check that number of chains matches number of tasks.
	chainCount, err := aggregators.CountChains()
	if err != nil {
		log.Printf("Chain count error: %v", err)
	} else {
		if chainCount != len(tasks) {
			log.Printf("Chain count error: Found %d chains but there are %d" +
					" tasks. To silence this error, place a nil placeholder" +
					" in the task list.", chainCount, len(tasks))
		}
	}
	
	// Perform aggregation tasks.
	t := time.Now()
	
	for _, task := range tasks {
		// A task may be nil, to make a placeholder for a future aggregator.
		if task == nil { continue }
	
		tt := time.Now()
		log.SetPrefix(task.name + " ")
		log.Printf("Starting %s.", task.name)
		
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
// sequentially. Use a nil value to make a placeholder, for chain counting.
var tasks = []*aggTask {
	&aggTask{ aggregators.Cerberus("TivTaam"), "TivTaam", "tivtaam" },
	nil,  // Placeholder for Co-Op.
	&aggTask{ aggregators.Shufersal(), "Shufersal", "shufersal" },
	&aggTask{ aggregators.Cerberus("DorAlon"), "DorAlon", "doralon" },
	&aggTask{ aggregators.Nibit(aggregators.Victory, 100), "Victory", "victory" },
	&aggTask{ aggregators.Nibit(aggregators.Hashook, 100), "Hashook", "hashook" },
	&aggTask{ aggregators.Nibit(aggregators.Lahav, 100), "Lahav", "lahav" },
	&aggTask{ aggregators.Cerberus("osherad"), "OsherAd", "osherad" },
	&aggTask{ aggregators.Mega(), "Mega", "mega" },
	&aggTask{ aggregators.Cerberus("HaziHinam"), "HaziHinam", "hazihinam" },
	&aggTask{ aggregators.Cerberus("Keshet"), "Keshet", "keshet" },
	&aggTask{ aggregators.Cerberus("RamiLevi"), "RamiLevi", "ramilevi" },
	&aggTask{ aggregators.Cerberus("SuperDosh"), "SuperDosh", "superdosh" },
	&aggTask{ aggregators.Cerberus("Yohananof"), "Yohananof", "yohananof" },
	&aggTask{ aggregators.Eden(), "Eden", "eden" },
	&aggTask{ aggregators.Bitan(), "Bitan", "bitan" },
	nil,  // Placeholder for Freshmarket.
}

// Returns the name that should be given to the log file.
func logFileName() string {
	t := time.Now()
	return fmt.Sprintf("Log-%d%02d%02d%02d%02d.txt",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

// Holds parsed command-line arguments.
var args struct {
	dir    string   // Where to download files.
	stdout *bool    // Log to stdout?
}

// Signifies that no args were given.
var noArgs = fmt.Errorf("")

// Parses arguments and places their values in the args struct. If an error
// returns, args are invalid.
func parseArgs() error {
	args.stdout = myflag.Bool("stdout", "", "Log to stdout instead of log" +
		" file.", false)
	
	err := myflag.Parse()
	if err != nil {
		return err
	}
	
	a := myflag.Args()
	
	// No args.
	if len(a) == 0 {
		return noArgs
	}
	
	// Too many args.
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

Flags:
`

// Prints a welcome message and usage instructions to the log.
func logWelcome() {
	log.Print("We have lift off!")
	
	// Print grep help.
	log.Print("To search for a specific chain use grep '^ChainName'.")
	log.Print("To search for errors, use grep 'error'.")
	log.Print("To search for times, use grep 'Took'.")
	
	// Print chain names.
	chains := ""
	for i := range tasks {
		if tasks[i] == nil { continue }
		if i > 0 { chains += ", " }
		chains += tasks[i].name
	}
	log.Print("Chains in this run: ", chains)
}

