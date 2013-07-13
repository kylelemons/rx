package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"
)

var helpCmd = &Command{
	Name:    "help",
	Usage:   "[<command>]",
	Summary: "Help on the rx command and subcommands.",
}

var helpDump = helpCmd.Flag.Bool("godoc", false, "Dump the godoc output for the command(s)")

func helpFunc(cmd *Command, args ...string) {
	var selected []*Command

	if len(args) > 0 {
		want := strings.ToLower(args[0])
		for _, cmd := range commands {
			if cmd.Name == want {
				selected = append(selected, cmd)
			}
		}
	}

	switch {
	case *helpDump:
		render(stdout, docTemplate, commands)
	case len(selected) < len(args):
		fmt.Fprintf(stdout, "error: unknown command %q\n", args[0])
		render(stdout, helpTemplate, helpCmd)
	case len(selected) == 0:
		render(stdout, usageTemplate, commands)
	case len(selected) == 1:
		render(stdout, helpTemplate, selected[0])
	}
}

func init() {
	helpCmd.Run = helpFunc
}

func tabify(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
}

var templateFuncs = template.FuncMap{
	"flags": func(indent int, args ...interface{}) string {
		b := new(bytes.Buffer)
		prefix := strings.Repeat(" ", indent)
		w := tabify(b)
		visit := func(f *flag.Flag) {
			dash := "--"
			if len(f.Name) == 1 {
				dash = "-"
			}
			eq := "= " + f.DefValue
			switch typeName := fmt.Sprintf("%T", f.Value); {
			case typeName == "*flag.stringValue":
				// TODO(kevlar): make my own stringValue type so as to not depend on this?
				eq = fmt.Sprintf("= %q", f.DefValue)
			case f.DefValue == "":
				eq = ""
			}
			fmt.Fprintf(w, "%s%s%s\t%s\t   %s\n", prefix, dash, f.Name, eq, f.Usage)
		}
		if len(args) == 0 {
			flag.VisitAll(visit)
		} else {
			args[0].(*Command).Flag.VisitAll(visit)
		}
		w.Flush()
		if b.Len() == 0 {
			return ""
		}
		return fmt.Sprintf("\nOptions:\n%s", b)
	},
	"title": func(s string) string {
		return strings.Title(s + " command")
	},
	"trim": func(s string) string {
		return strings.TrimSpace(s)
	},
}

var stdout io.Writer = tabConverter{os.Stdout}

type tabConverter struct{ io.Writer }

func (t tabConverter) Write(p []byte) (int, error) {
	p = bytes.Replace(p, []byte{'\t'}, []byte{' ', ' ', ' ', ' '}, -1)
	return t.Writer.Write(p)
}

func render(w io.Writer, tpl string, data interface{}) {
	t := template.New("help")
	t.Funcs(templateFuncs)
	if err := template.Must(t.Parse(tpl)).Execute(w, data); err != nil {
		panic(err)
	}
}

var generalHelp = `	rx [<options>] [<subcommand> [<suboptions>] [<arguments> ...]]
{{flags 2}}
Commands:{{range .}}
	{{.Name | printf "%-10s"}} {{.Summary}}{{end}}

Use "rx help <command>" for more help with a command.
`

var usageTemplate = `rx is a command-line dependency management tool for Go projects.

Usage:
` + generalHelp

var helpTemplate = `Usage: rx {{.Name}} [options]{{with .Usage}} {{.}}{{end}}{{if .Abbrev}}
       rx {{.Abbrev}} [options]{{with .Usage}} {{.}}{{end}}
{{end}}{{flags 2 .}}
{{.Summary}}
{{if .Help}}
{{.Help | trim}}{{end}}
`

var docTemplate = `/*
The rx command is a dependency and version management system for Go projects.
It is built on top of the go tool and utilizes the $GOPATH convention.

Installation

As usual, the rx tool can be installed or upgraded via the "go" tool:
	go get -u kylelemons.net/go/rx

General Usage

The rx command is composed of numerous sub-commands.
Sub-commands can be abbreviated to any unique prefix on the command-line.
The general usage is:

` + generalHelp + `

See below for a description of the various sub-commands understood by rx.
{{range .}}
{{.Name | title}}

{{.Summary | trim}}

Usage:
	rx {{.Name}} {{.Usage | trim}}
{{flags 2 .}}
{{.Help | trim}}
{{end}}
*/
package main
`
