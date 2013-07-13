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
	cpointCmd,
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

	var found []*Command
	sub, args := args[0], args[1:]
find:
	for _, cmd := range commands {
		if sub == cmd.Abbrev {
			found = []*Command{cmd}
			break find
		}
		if strings.HasPrefix(cmd.Name, sub) {
			found = append(found, cmd)
		}
	}
	// Scan first (this is a no-op unless load failed or --rescan)
	if err := Scan(); err != nil {
		fmt.Fprintf(stdout, "error: scan: %s", err)
		os.Exit(1)
	}

	switch cnt := len(found); cnt {
	case 1:
		found[0].Exec(args)
	case 0:
		fmt.Fprintf(stdout, "error: unknown command %q\n\n", sub)
		flag.Usage()
		os.Exit(1)
	default:
		fmt.Fprintf(stdout, "error: non-unique command prefix %q (matched %d commands)\n\n", sub, cnt)
		flag.Usage()
		os.Exit(1)
	}
}
