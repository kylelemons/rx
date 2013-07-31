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
	"kylelemons.net/go/rx/graph"
)

var tagsCmd = &Command{
	Name:    "tags",
	Usage:   "<repo>",
	Summary: "List known repository tags.",
	Help: `The tags command scans the specified repository and lists
information about its tags.  The <repo> can be any piece of the repository root
path, as long as it is unique.

The -f option takes a template as a format.  The data passed into the
template invocation is an (rx/graph) TagList, and the default format is:

` + ind2sp(tagsTemplate),
}

var (
	tagsFormat = tagsCmd.Flag.String("f", "", "tags output format")
	tagsLong   = tagsCmd.Flag.Bool("long", false, "Use long output format")
	tagsUp     = tagsCmd.Flag.Bool("up", false, "Only show updates (overrides --down)")
	tagsDown   = tagsCmd.Flag.Bool("down", false, "Only show downgrades")
)

func tagsFunc(cmd *Command, args ...string) {
	switch len(args) {
	case 0:
		args = append(args, "all")
	case 1:
	default:
		cmd.BadArgs("too many arguments")
	}
	path := args[0]

	repo, err := Deps.FindRepo(path)
	if err != nil {
		cmd.Fatalf("<repo>: %s", err)
	}

	var tags graph.TagList
	switch {
	case *tagsUp:
		tags, err = repo.Upgrades()
	case *tagsDown:
		tags, err = repo.Downgrades()
	default:
		tags, err = repo.Tags()
	}
	if err != nil {
		cmd.Fatalf("list tags for %q: %s", repo.Root, err)
	}

	switch {
	case *tagsFormat != "":
		render(stdout, *tagsFormat, tags)
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
)
