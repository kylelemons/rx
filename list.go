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

` + ind2sp(listTemplate) + `

If you specify --long, the format will be:

` + ind2sp(listTemplateLong),
}

var (
	listFormat = listCmd.Flag.String("f", "", "List output format")
	listLong   = listCmd.Flag.Bool("long", false, "Use long output format")
)

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

	switch {
	case *listFormat != "":
		render(stdout, *listFormat, Repos)
	case *listLong:
		render(stdout, listTemplateLong, Repos)
	default:
		render(stdout, listTemplate, Repos)
	}
}

func init() {
	listCmd.Run = listFunc
}

var (
	listTemplate = `{{range .}}{{.Path}}:{{range .Packages}} {{.Name}}{{end}}
{{end}}`

	listTemplateLong = `{{range .}}Repository ({{.VCS}}) {{printf "%q" .Path}}:
	Dependencies:{{range .RepoDeps}}
		{{.}}{{end}}
	Packages:{{range .Packages}}
		{{.ImportPath}}{{end}}

{{end}}`
)

func ind2sp(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}
