// Handy functionality for flag parsing.
package myflag

import (
	"os"
	"fmt"
	"flag"
	"bytes"
)

func init() {
	Reset()
}

// Holds flags.
var flags *flag.FlagSet

// Instanciates the flag set.
func Reset() {
	flags = flag.NewFlagSet("", flag.ContinueOnError)
}

// Parses the program's input arguments.
func Parse() error {
	return ParseStrings(os.Args[1:])
}

// Parses the given slice of strings.
func ParseStrings(args []string) error {
	if len(args) == 0 {
		hasAny = false
	} else {
		hasAny = true
	}
	
	return flags.Parse(args)
}

// Returns non-flag arguments.
func Args() []string {
	return flags.Args()
}

// Determines if ANY flags were given to the parse function.
var hasAny bool

// Determines if ANY flags were given to the parse function.
// Useful for help message printing.
func HasAny() bool {
	return hasAny
}


// ----- FLAG REGISTERING ------------------------------------------------------

// Registers a new int flag.
func Int(name string, shortName string, typ string, description string,
		value int) *int {
	// Register long name
	result := flags.Int(name, value, "")
	
	// Register short name
	if shortName != "" {
		flags.IntVar(result, shortName, value, "")
	}
	
	// Add help message
	addHelpMessage(name, shortName, typ, description)
	
	return result
}

// Registers a new float flag.
func Float(name string, shortName string, typ string, description string,
		value float64) *float64 {
	// Register long name
	result := flags.Float64(name, value, "")
	
	// Register short name
	if shortName != "" {
		flags.Float64Var(result, shortName, value, "")
	}
	
	// Add help message
	addHelpMessage(name, shortName, typ, description)
	
	return result
}

// Registers a new string flag.
func String(name string, shortName string, typ string, description string,
		value string) *string {
	// Register long name
	result := flags.String(name, value, "")
	
	// Register short name
	if shortName != "" {
		flags.StringVar(result, shortName, value, "")
	}
	
	// Add help message
	addHelpMessage(name, shortName, typ, description)
	
	return result
}

// Registers a new boolean flag.
func Bool(name string, shortName string, description string,
		value bool) *bool {
	// Register long name
	result := flags.Bool(name, value, "")
	
	// Register short name
	if shortName != "" {
		flags.BoolVar(result, shortName, value, "")
	}
	
	// Add help message
	addHelpMessage(name, shortName, "", description)
	
	return result
}


// ----- HELP STRING -----------------------------------------------------------

// Help message for the flags. Accumulates help messages as flags are set.
var flagsHelp = bytes.NewBuffer(nil)

// Returns a pretty string representation of the flags.
func Help() string {
	return flagsHelp.String()
}

// Adds a help message for the given flag.
func addHelpMessage(name string, shortName string, typ string,
		description string) {
	// Modify type to fit printing
	if typ != "" {
		typ = " <" + typ + ">"
	}
	
	// Add help message
	if shortName != "" {
		fmt.Fprintf(flagsHelp, "\t-%s%s\n", shortName, typ)
	}
	fmt.Fprintf(flagsHelp, "\t-%s%s\n", name, typ)
	fmt.Fprintf(flagsHelp, "\t\t%s\n\n", description)
}

