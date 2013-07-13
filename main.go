package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	verbose = flag.Bool("v", false, "Turn on verbose logging")
)

var commands = []*Command{
	helpCmd,
	listCmd,
	tagsCmd,
	preCmd,
	cabCmd,
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

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	Load()
	defer Save()

	var found *Command
	sub, args := args[0], args[1:]
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.Name, sub) {
			if found != nil {
				fmt.Fprintf(stdout, "error: non-unique command prefix %q\n\n", sub)
				os.Exit(1)
			}
			found = cmd
		}
	}

	// Scan first (this is a no-op unless load failed or --rescan)
	if err := Scan(); err != nil {
		fmt.Fprintf(stdout, "error: scan: %s", err)
		os.Exit(1)
	}

	if found == nil {
		fmt.Fprintf(stdout, "error: unknown command %q\n\n", sub)
		flag.Usage()
		os.Exit(1)
	}
	found.Exec(args)
}
