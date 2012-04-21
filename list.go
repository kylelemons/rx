package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"sort"
)

var listCmd = &Command{
	Name:    "list",
	Usage:   "[package|dir]",
	Summary: "List recognized repositories",
	Run:     listFunc,
}
func listFunc(cmd *Command, args ...string) {
	switch len(args) {
	case 0:
		args = append(args, "all")
	case 1:
	default:
		cmd.BadArgs("too many arguments")
	}

	list := exec.Command("go", "list", "-json")
	list.Args = append(list.Args, args...)
	list.Stderr = stdout
	js, err := list.Output()
	if err != nil {
		cmd.Fatalf("go list: %s", err)
	}
	dec := json.NewDecoder(bytes.NewReader(js))

	repoPkgs := map[string][]*Package{}
	repoVCS := map[string]string{}
	pkgRoot := map[string]string{}
	for {
		pkg := new(Package)

		// Decode the next package
		err := dec.Decode(pkg)
		if err == io.EOF {
			break
		} else if err != nil {
			cmd.Fatalf("parsing package: %s", err)
		}

		// Only save packages we want to keep
		if !pkg.Keep() {
			continue
		}

		// Detect the version control system
		vcs, root := pkg.DetectVCS()

		// Ignore things we don't understand :D
		if vcs == "" {
			continue
		}

		repoPkgs[root] = append(repoPkgs[root], pkg)
		repoVCS[root] = vcs
		pkgRoot[pkg.ImportPath] = root
	}

	repos := map[string]*Repository{}

	// Create the repositories
	for path := range repoVCS {
		r := &Repository{
			Path: path,
			VCS:  repoVCS[path],
			Packages: repoPkgs[path],
		}

		// Find repo dependencies
		deps := map[string]bool{}
		for _, pkg := range r.Packages {
			for _, dep := range pkg.Imports {
				if depRepo, ok := pkgRoot[dep]; ok {
					deps[depRepo] = true
				}
			}
		}
		for dep := range deps {
			r.RepoDeps = append(r.RepoDeps, dep)
		}
		sort.Strings(r.RepoDeps)

		repos[path] = r
	}

	render(stdout, repoTemplate, repos)
}
