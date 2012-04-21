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
	Usage:   "[command]",
	Summary: "Help on the rx command and subcommands",
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

var usageTemplate = `rx is a command-line dependency management tool for Go projects.

Usage:
	rx <command> [arguments]
{{flags 2}}
Commands:{{range .}}
	{{.Name | printf "%-10s"}} {{.Summary}}{{end}}

Use "rx help <command>" for more help with a command.
`

var helpTemplate = `Usage: rx {{.Name}} [options] {{.Usage}}
{{flags 2 .}}
{{.Summary}}
{{if .Help}}
{{.Help}}{{end}}
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

var templateFuncs = template.FuncMap{
	"flags": func(indent int, args ...interface{}) string {
		b := new(bytes.Buffer)
		prefix := strings.Repeat(" ", indent)
		w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)
		visit := func(f *flag.Flag) {
			dash := "--"
			if len(f.Name) == 1 {
				dash = "-"
			}
			fmt.Fprintf(w, "%s%s%s\t=\t%#v\t   %s\n", prefix, dash, f.Name, f.DefValue, f.Usage)
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
