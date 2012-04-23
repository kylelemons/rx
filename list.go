package main

import (
	"strings"
	"regexp"

	"kylelemons.net/go/rx/graph"
)

var listCmd = &Command{
	Name:    "list",
	Usage:   "[<filter>]",
	Summary: "List recognized repositories.",
	Help: `The list command scans all available packages and collects information about
their repositories.  By default, each repository is listed along with its
dependencies and contained packages. If a <filter> regular expression is
provided, only repositories whose root path matches the filter will be listed.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/graph) Graph, and the default format is:

` + ind2sp(listTemplate) + `

If you specify --long, the format will be:

` + ind2sp(listTemplateLong),
}

var (
	listFormat = listCmd.Flag.String("f", "", "List output format")
	listLong   = listCmd.Flag.Bool("long", false, "Use long output format")
)

func listFunc(cmd *Command, args ...string) {
	data := struct {
		Repository map[string]*graph.Repository
		*graph.Graph
	}{
		Repository: Deps.Repository,
		Graph:      Deps,
	}

	switch len(args) {
	case 0:
	case 1:
		filter, err := regexp.Compile(args[0])
		if err != nil {
			cmd.BadArgs("<filter> failed to compile: %s", err)
		}

		data.Repository = make(map[string]*graph.Repository)
		for path, repo := range Deps.Repository {
			if filter.MatchString(path) {
				data.Repository[path] = repo
			}
		}
	default:
		cmd.BadArgs("too many arguments")
	}

	switch {
	case *listFormat != "":
		render(stdout, *listFormat, data)
	case *listLong:
		render(stdout, listTemplateLong, data)
	default:
		render(stdout, listTemplate, data)
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
