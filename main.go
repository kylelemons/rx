package main

import (
	"flag"
)

var commands = []*Command{
	helpCmd,
}

func main() {
	flag.Usage = func() {
		helpRun(helpCmd)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
	}
}
