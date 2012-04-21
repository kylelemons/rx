package main

import (
	"flag"
	"fmt"
)

var commands = []*Command{
	helpCmd,
	listCmd,
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

	sub, args := args[0], args[1:]
	for _, cmd := range commands {
		if cmd.Name == sub {
			cmd.Exec(args)
			return
		}
	}

	fmt.Fprintf(stdout, "error: unknown command %q\n\n", sub)
	flag.Usage()
}
