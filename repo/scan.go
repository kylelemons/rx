package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
)

type RepoMap map[string]*Repository

func Scan() (RepoMap, error) {
	list := exec.Command("go", "list", "-json", "all")
	js, err := list.Output()
	if err != nil {
		return nil, fmt.Errorf("repo: go list: %s", err)
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
			return nil, fmt.Errorf("repo: parsing package: %s", err)
		}

		// Only save packages we want to keep
		if !pkg.Keep() {
			log.Printf("Skipping %q", pkg.ImportPath)
			continue
		}
		log.Printf("Adding %q", pkg.ImportPath)

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

	repos := RepoMap{}

	// Create the repositories
	for path := range repoVCS {
		r := &Repository{
			Path:     path,
			VCS:      repoVCS[path],
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
			for _, dep := range pkg.TestImports {
				if depRepo, ok := pkgRoot[dep]; ok {
					deps[depRepo] = true
				}
			}
			// TODO(kevlar): XTestImports?
		}
		for dep := range deps {
			r.RepoDeps = append(r.RepoDeps, dep)
		}
		sort.Strings(r.RepoDeps)

		repos[path] = r
	}

	return repos, nil
}
