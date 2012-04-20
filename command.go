package main

import (
	"fmt"
	"bytes"
	"strings"
	"text/tabwriter"
	"flag"
)

type Command struct {
	// This function is called when the command is invoked
	Run   func(cmd *Command, args ...string)

	// Command-line flags
	Flag flag.FlagSet

	Name  string // The name of the command
	Usage string // The symbolic, human-readable argument description

	Summary string // The short description of the command
	Help    string // The detailed command information
}

func (c *Command) Exec(args []string) {
	c.Flag.Usage = func() {
		helpRun(c, c.Name)
	}
	c.Flag.Parse(args)
	c.Run(c, c.Flag.Args()...)
}

func (c *Command) FlagDump(indent int) string {
	b := new(bytes.Buffer)
	prefix := strings.Repeat(" ", indent)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)
	c.Flag.VisitAll(func(f *flag.Flag){
		fmt.Fprintf(w, "%s--%s\t=\t%v\t   %s\n", prefix, f.Name, f.DefValue, f.Usage)
	})
	w.Flush()
	return b.String()
}
