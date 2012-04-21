package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

var helpCmd = &Command{
	Name: "help",
	Usage: "[command]",
	Summary: "Help on the rx command and subcommands",
}

var helpDump = helpCmd.Flag.Bool("godoc", false, "Dump the godoc output for the command(s)")

func helpRun(cmd *Command, args ...string) {
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
	helpCmd.Run = helpRun
}

var usageTemplate = `rx is a command-line dependency management tool for Go projects.

Usage:
  rx <command> [arguments]

Commands:{{range .}}
  {{.Name | printf "%-10s"}} {{.Summary}}{{end}}

Use "rx help <command>" for more help with a command.
`

var helpTemplate = `Usage: rx {{.Name}} [options] {{.Usage}}
{{.FlagDump 2}}
{{.Summary}}
{{.Help}}
`

var docTemplate = `/*
{{range .}}Command: rx {{.Name}}

{{.Summary}}

Usage:
  rx {{.Usage}}

{{.Help}}{{end}}
*/
package documentation
`

var stdout io.Writer = os.Stdout
func render(w io.Writer, tpl string, data interface{}) {
	if err := template.Must(template.New("help").Parse(tpl)).Execute(w, data); err != nil {
		panic(err)
	}
}
