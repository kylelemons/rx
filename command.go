package main

import (
	"flag"
	"fmt"
	"os"
)

// A Command represents a subcommand of rx.
//
// The specified fields should be specified in the Command literal, usually at
// the top of the file.  Due to circular dependencies, the Run command will
// typically have to be set in an init if it uses flags defined by the command.
//
// The abbreviation should be used for an abbreviation that is not a strict
// prefix of the command name, as those work automatically.  Most commands will
// not need one, as the prefixes will work well enough.
type Command struct {
	// Run is called when the command is invoked
	Run func(cmd *Command, args ...string)

	// Flag contains Command-line flags
	Flag flag.FlagSet

	// These should be set in the literal:
	Name    string // The name of the command (all lower case, one word)
	Usage   string // The symbolic, human-readable argument description
	Summary string // The short description of the command (short sentence)
	Help    string // The detailed command information (multiple paragraphs, etc)
	Abbrev  string // Alternate abbreviation for the command

	exit int // Exit code
}

func (c *Command) Exec(args []string) {
	c.Flag.Usage = func() {
		helpFunc(c, c.Name)
	}
	c.Flag.Parse(args)

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(fatal); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	c.Run(c, c.Flag.Args()...)

	if c.exit != 0 {
		os.Exit(c.exit)
	}
}

func (c *Command) BadArgs(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, "error: "+errFormat+"\n\n", args...)
	helpFunc(c, c.Name)
	panic(fatal{})
}

// Errorf prints out a formatted error with the right prefixes.
func (c *Command) Errorf(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, c.Name+": error: "+errFormat+"\n", args...)
	if c.exit == 0 {
		c.exit = 1
	}
}

// Fatalf is like Errorf except the stack unwinds up to the Exec call before
// exiting the application with status code 1.
func (c *Command) Fatalf(errFormat string, args ...interface{}) {
	c.Errorf(errFormat, args...)
	panic(fatal{})
}

type fatal struct{}
