// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
