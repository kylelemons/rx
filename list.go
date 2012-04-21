package main

import (
	"strings"
)

var listCmd = &Command{
	Name:    "list",
	Summary: "List recognized repositories",
	Help: `The list command scans all available packages and collects information about
their repositories.  By default, each repository is listed along with its
dependencies and contained packages.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/repo) RepoMap, and the default format is:

` + ind2sp(listTemplate),
}

var listFormat = listCmd.Flag.String("f", "", "List output format")

func listFunc(cmd *Command, args ...string) {
	switch len(args) {
	case 0:
		args = append(args, "all")
	case 1:
	default:
		cmd.BadArgs("too many arguments")
	}

	// Scan before accessing Repos
	if err := Scan(); err != nil {
		cmd.Fatalf("scan: %s", err)
	}

	if *listFormat != "" {
		render(stdout, *listFormat, Repos)
		return
	}

	render(stdout, listTemplate, Repos)
}

func init() {
	listCmd.Run = listFunc
}

var listTemplate = `{{range .}}Repository ({{.VCS}}) {{printf "%q" .Path}}:
	Dependencies:{{range .RepoDeps}}
		{{.}}{{end}}
	Packages:{{range .Packages}}
		{{.ImportPath}}{{end}}

{{end}}`

func ind2sp(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}
