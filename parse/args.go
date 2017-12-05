package main

// Handles command line flags.

import (
	"flag"
	"runtime"
	"os"

	"github.com/fluhus/gostuff/flug"
)

var args struct {
	Files    []string
	Check    bool   `flug:"c,Only check input files, do not create output tables."`
	OutDir   string `flug:"o,Output directory. Default is current."`
	ForceRaw bool   `flug:"f,Force parsing of raw files, instead of reading serialized data."`
	NumThreads int `flug:"t,Number of threads to run on. Default is number of CPUs."`
	Help     bool
}

// TODO(amit): Expand file arguments to a full, sorted input file list.

func parseArgs() {
	flag.Usage = printArgError
	err := flug.Register(&args)
	if err != nil {
		panic(err)
	}

	if len(os.Args) == 1 {
		printUsage()
		os.Exit(1)
	}

	// Parse flags.
	flag.Parse()

	if args.NumThreads == 0 {
		args.NumThreads = runtime.NumCPU()
	}

	args.Files = flag.Args()
	if len(args.Files) == 0 {
		pe("No input files provided.")
		printArgError()
		os.Exit(1)
	}
}

func printArgError() {
	pe("Run without arguments for help.")
}

func printUsage() {
	pe(help)
	flag.PrintDefaults()
	pe()
	pe(credit)
}

// Help message to display.
const help = `Parses XML files for the supermarket price project.

Outputs TSV text files to the output directory. Supports XML, ZIP and GZ
formats. Also generates for each input file an intermediate data file with
the '.items' suffix. DO NOT USE THESE FILES AS INPUT. Use the standard data
files and the program will automatically read the intermediate if it can.

Usage:
parse [OPTIONS] file/dir1 file/dir2 file/dir3 ...

Arguments:`

const credit = `Credit:
Based on the 'prices' project by Amit Lavon.
https://github.com/fluhus/prices`
