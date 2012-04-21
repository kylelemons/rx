package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var commands = []*Command{
	helpCmd,
	listCmd,
	tagsCmd,
}

func main() {
	flag.Usage = func() {
		helpFunc(helpCmd)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	Load()
	defer Save()

	var found *Command
	sub, args := args[0], args[1:]
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Name, sub) {
			if found != nil {
				fmt.Fprintf(stdout, "error: non-unique prefix %q\n\n", sub)
				os.Exit(1)
			}
			found = cmd
		}
	}
	if found == nil {
		fmt.Fprintf(stdout, "error: unknown command %q\n\n", sub)
		flag.Usage()
		os.Exit(1)
	}
	found.Exec(args)
}
