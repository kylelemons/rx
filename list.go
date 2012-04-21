package main

import (
	"github.com/kylelemons/rx/repo"
)

var listCmd = &Command{
	Name:    "list",
	Usage:   "[package|dir]",
	Summary: "List recognized repositories",
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

	repos, err := repo.Scan()
	if err != nil {
		cmd.Fatalf("scan: %s", err)
	}

	if *listFormat != "" {
		render(stdout, *listFormat, repos)
		return
	}

	render(stdout, listTemplate, repos)
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
