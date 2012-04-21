package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	rescan = flag.Bool("rescan", false, "Force a rescan of repositories")
)

type Command struct {
	// This function is called when the command is invoked
	Run func(cmd *Command, args ...string)

	// Command-line flags
	Flag flag.FlagSet

	Name  string // The name of the command
	Usage string // The symbolic, human-readable argument description

	Summary string // The short description of the command
	Help    string // The detailed command information
}

func (c *Command) Exec(args []string) {
	c.Flag.Usage = func() {
		helpFunc(c, c.Name)
	}
	c.Flag.Parse(args)
	c.Run(c, c.Flag.Args()...)
}

func (c *Command) BadArgs(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, "error: "+errFormat+"\n\n", args...)
	helpFunc(c, c.Name)
	os.Exit(1)
}

func (c *Command) Fatalf(errFormat string, args ...interface{}) {
	fmt.Fprintf(stdout, c.Name+": error: "+errFormat, args...)
	os.Exit(1)
}
