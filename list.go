package main

import (
	"strings"
)

var listCmd = &Command{
	Name:    "list",
	Summary: "List recognized repositories.",
	Help: `The list command scans all available packages and collects information about
their repositories.  By default, each repository is listed along with its
dependencies and contained packages.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/graph) RepoMap, and the default format is:

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

	// Scan before accessing Deps
	if err := Scan(); err != nil {
		cmd.Fatalf("scan: %s", err)
	}

	switch {
	case *listFormat != "":
		render(stdout, *listFormat, Deps)
	case *listLong:
		render(stdout, listTemplateLong, Deps)
	default:
		render(stdout, listTemplate, Deps)
	}
}

func init() {
	listCmd.Run = listFunc
}

var (
	listTemplate = `{{range .Repository}}{{.}} :{{range .Packages}}{{$pkg := index $.Package .}} {{$pkg.Name}}{{end}}
{{end}}`

	listTemplateLong = `{{range .Repository}}Repository ({{.VCS}}) {{.}}:
	Packages:{{range .Packages}}
		{{$pkg := index $.Package .}}{{$pkg.ImportPath}}{{end}}
{{with $.RepoDeps .}}	Dependencies:{{range .}}
		{{.}}{{end}}
{{end}}{{with $.RepoUsers .}}	Users:{{range .}}
		{{.}}{{end}}
{{end}}
{{end}}`
)

func ind2sp(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "  " + lines[i]
	}
	return strings.Join(lines, "\n")
}
