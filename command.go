package main

import (
	"flag"
	"fmt"
	"os"
)

type Command struct {
	// This function is called when the command is invoked
	Run func(cmd *Command, args ...string)

	// Command-line flags
	Flag flag.FlagSet

	Name  string // The name of the command (all lower case, one word)
	Usage string // The symbolic, human-readable argument description

	Summary string // The short description of the command (short sentence)
	Help    string // The detailed command information (multiple paragraphs, etc)
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
}

func (c *Command) BadArgs(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, "error: "+errFormat+"\n\n", args...)
	helpFunc(c, c.Name)
	panic(fatal{})
}

// Errorf prints out a formatted error with the right prefixes.
func (c *Command) Errorf(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, c.Name+": error: "+errFormat+"\n", args...)
}

// Fatalf is like Errorf except the stack unwinds up to the Exec call before
// exiting the application with status code 1.
func (c *Command) Fatalf(errFormat string, args ...interface{}) {
	c.Errorf(errFormat, args...)
	panic(fatal{})
}

type fatal struct{}
