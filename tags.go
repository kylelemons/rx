package main

import (
	"strings"

	"github.com/kylelemons/rx/repo"
)

var tagsCmd = &Command{
	Name:    "tags",
	Usage:   "repo",
	Summary: "list repository tags",
	Help: `The tags command scans the specified repository and lists
information about its tags.  The [repo] can be the suffix of the repository
root path, as long as it is unique.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/repo) TagList, and the default format is:

` + ind2sp(tagsTemplate) + `

If you specify --long, the format will be:

` + ind2sp(tagsTemplateLong),
}

var (
	tagsFormat = tagsCmd.Flag.String("f", "", "tags output format")
	tagsLong = tagsCmd.Flag.Bool("long", false, "Use long output format")
)

func tagsFunc(cmd *Command, args ...string) {
	switch len(args) {
	case 0:
		args = append(args, "all")
	case 1:
	default:
		cmd.BadArgs("too many arguments")
	}
	pathSuffix := args[0]

	// Scan before accessing Repos
	if err := Scan(); err != nil {
		cmd.Fatalf("scan: %s", err)
	}

	var tags repo.TagList
	for path, repo := range Repos {
		if strings.HasSuffix(path, pathSuffix) {
			if tags != nil {
				cmd.Fatalf("non-unique suffix %q", pathSuffix)
			}

			var err error
			tags, err = repo.Tags()
			if err != nil {
				cmd.Fatalf("list tags for %q: %s", path, err)
			}
		}
	}
	if tags == nil {
		cmd.Fatalf("unknown repo %q", pathSuffix)
	}

	switch {
	case *tagsFormat != "":
		render(stdout, *tagsFormat, tags)
	case *tagsLong:
		render(stdout, tagsTemplateLong, tags)
	default:
		render(stdout, tagsTemplate, tags)
	}
}

func init() {
	tagsCmd.Run = tagsFunc
}

var (
	tagsTemplate = `{{range .}}{{.Rev}} {{.Name}}
{{end}}`

	tagsTemplateLong =``
)