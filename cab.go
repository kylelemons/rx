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
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"kylelemons.net/go/rx/graph"
)

var cabCmd = &Command{
	Name:    "cabinet",
	Usage:   "<repo> [<id>]",
	Summary: "Save, list, or restore dependency snapshots.",
	Help: `The cabinet command saves dependency information for the given
repository in a file within it.  The <repo> can be any piece of the
repository root path, as long as it is unique.  The <tag> is anything
understood by the underlying version control system as a commit, usually a
tag, branch, or commit.  When saving, the <id> will override the default
(based on the date) and when opening the <id> is the ID (or a unique
substring) of the cabinet to open.

Cabinets are currently stored in the repository itself named as follows:
	.rx/cabinet-YYYYMMDD-HHMMSS

If the <id> is specified for --build, it will override the date-based portion
of the cabinet name.  If a cabinet already exists with the name (even if it is
a date-based name) it will not be overwritten.

Unless --test=false is specified, the packages in the repository will be tested
before a cabinet is created and after a cabinet is restored.  If the test
before creation fails, the cabinet will not be created.  If the test after
restoration fails, the repositories will be reverted to their original
revisions.

The cabinet format is still in flux.`,
}

const cabIDFormat = "20060102-150405"

var (
	cabTest  = cabCmd.Flag.Bool("test", true, "test package before saving and after loading cabinet")
	cabList  = cabCmd.Flag.Bool("list", true, "list matching cabinet files (the default)")
	cabBuild = cabCmd.Flag.Bool("build", false, "create a new cabinet")
	cabOpen  = cabCmd.Flag.Bool("open", false, "open the specified cabinet")
	cabDump  = cabCmd.Flag.Bool("dump", false, "list the contents of the specified cabinet")
)

func cabFunc(cmd *Command, args ...string) {
	if len(args) < 1 || len(args) > 2 {
		cmd.BadArgs("requires two arguments")
	}
	path := args[0]

	repo, err := Deps.FindRepo(path)
	if err != nil {
		cmd.Fatalf("<repo>: %s", err)
	}

	var id string
	if len(args) > 1 {
		id = args[1]
	}

	switch {
	case *cabBuild:
		if id == "" {
			id = time.Now().Format(cabIDFormat)
		}
		err = buildCabinet(repo, id)
	case *cabOpen:
		err = openCabinet(cmd, repo, id)
	case *cabDump:
		if id == "" {
			cmd.BadArgs("must specify <id> to dump")
		}
		err = dumpCabinet(repo, id)
	case *cabList:
		err = listCabinets(repo, id)
	default:
		cmd.BadArgs("no mode specified")
	}
	if err != nil {
		cmd.Fatalf("%s", err)
	}
}

// A CabFile is the data structure stored in a cabinet file.
type CabFile struct {
	Repo    string         // The import path prefix covered by this cabinet (usually import/path/...)
	Created time.Time      // The time the cabinet was created
	Head    string         // The hash of the repository at which this cabinet was created
	Deps    []*RepoVersion // The dependencies of this package
}

// A RepoVersion stores information about a dependency of a repository.
type RepoVersion struct {
	Pattern  string   // The pattern required to scan for updates
	Packages []string // Try `go get -d` on these in order until one succeeds
	Head     string   // The hash of the repository to use after installation
}

// NewRepoVersion creates a repo version object suitable for storing into cabinets, etc.
func NewRepoVersion(repo *graph.Repository) (*RepoVersion, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get %s head: %s", repo, err)
	}
	return &RepoVersion{
		Pattern:  repo.String(),
		Packages: repo.Packages,
		Head:     head,
	}, nil
}

// Apply attempts to locate the repository and pin it to the head version.
func (dep *RepoVersion) Apply() error {
	// Find the repository
	var repo *graph.Repository
	for _, pkg := range dep.Packages {
		// Make sure the package is installed and up-to-date
		get := exec.Command("go", "get", "-d", pkg)
		get.Dir = os.TempDir()
		get.Stdout = os.Stdout
		get.Stderr = os.Stderr
		if err := get.Run(); err != nil {
			continue
		}

		// Scan for new packages
		if err := Deps.Scan(dep.Pattern); err != nil {
			continue
		}

		// See if we found the package we wanted
		p, ok := Deps.Package[pkg]
		if !ok {
			continue
		}

		// Get the repository
		r, ok := Deps.Repository[p.RepoRoot]
		if !ok {
			continue
		}

		repo = r
		break
	}
	if repo == nil {
		return fmt.Errorf("apply(%q@%q): unable to locate repository", dep.Pattern, dep.Head)
	}

	// Get fallback in case the update fails
	fallback, err := repo.Head()
	if err != nil {
		return fmt.Errorf("apply(%q@%q): unable to determine fallback version", dep.Pattern, dep.Head)
	}

	// Pin the version
	if err := repo.ToRev(dep.Head); err != nil {
		if ferr := repo.ToRev(fallback); ferr != nil {
			return fmt.Errorf("apply(%q): pin(%q) [%s] and fallback(%q) [%s] failed",
				dep.Pattern, dep.Head, err, fallback, ferr)
		}
		return fmt.Errorf("apply(%q@%q): pin failed: %s", dep.Pattern, dep.Head, err)
	}

	log.Printf("Pinned %s @ %s", dep.Pattern, dep.Head)
	return nil
}

func buildCabinet(repo *graph.Repository, id string) error {
	// Build the cabinet data
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("build: get repo head: %s", err)
	}
	data := &CabFile{
		Repo:    repo.String(),
		Created: time.Now(),
		Head:    head,
	}
	deps, err := Deps.RepoDepTree(repo)
	if err != nil {
		return fmt.Errorf("build: scan dependencies: %s", err)
	}
	for _, dep := range deps {
		rv, err := NewRepoVersion(dep)
		if err != nil {
			return fmt.Errorf("build: %s", err)
		}
		data.Deps = append(data.Deps, rv)
	}

	if *cabTest {
		// Test the packages in the repository
		test := exec.Command("go", "test", repo.String())
		test.Dir = os.TempDir()
		test.Stdout = os.Stdout
		test.Stderr = os.Stderr
		if err := test.Run(); err != nil {
			return fmt.Errorf("build: `go test` failed: %s", err)
		}
	}

	// Make the $REPO/.rx directory
	repoRx := filepath.Join(repo.Root, ".rx")
	if err := os.MkdirAll(repoRx, 0755); err != nil {
		return fmt.Errorf("build: create .rx: %s", err)
	}

	filename := filepath.Join(repoRx, "cabinet-"+id)

	// Open the file, but fail if it already exists
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("build: open cabinet: %s", err)
	}
	defer file.Close()

	if err := gob.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("build: encoding cabinet: %s", err)
	}

	log.Printf("Cabinet written to %q", filename)
	return nil
}

func listCabinetFiles(repo *graph.Repository, filter string) ([]string, error) {
	pattern := filepath.Join(repo.Root, ".rx", "cabinet-*")

	all, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if filter == "" {
		return all, nil
	}

	re, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, filename := range all {
		if !re.MatchString(filename) {
			continue
		}
		matches = append(matches, filename)
	}
	return matches, nil
}

func loadCabinet(repo *graph.Repository, id string) (string, *CabFile, error) {
	files, err := listCabinetFiles(repo, id)
	if err != nil {
		return "", nil, fmt.Errorf("list: %s", err)
	}
	switch cnt := len(files); {
	case cnt == 0:
		return "", nil, fmt.Errorf("no matching cabinet files found")
	case cnt > 1:
		return "", nil, fmt.Errorf("non-unique id pattern %q (matched %d cabinets)", id, cnt)
	}
	filename := files[0]

	file, err := os.Open(filename)
	if err != nil {
		return filename, nil, fmt.Errorf("open cabinet: %s", err)
	}

	data := new(CabFile)
	if err := gob.NewDecoder(file).Decode(data); err != nil {
		return filename, nil, fmt.Errorf("decode cabinet: %s", err)
	}
	return filename, data, nil
}

func openCabinet(cmd *Command, repo *graph.Repository, id string) error {
	filename, data, err := loadCabinet(repo, id)
	if err != nil {
		return fmt.Errorf("open: %s", err)
	}

	var errors int
	for _, dep := range data.Deps {
		if err := dep.Apply(); err != nil {
			errors++
			cmd.Errorf("open: %s", err)
		}
	}
	if errors > 0 {
		return fmt.Errorf("open: %d repositories could not be pinned", errors)
	}

	log.Printf("Opened cabinet %q", filename)
	return nil
}

func listCabinets(repo *graph.Repository, id string) error {
	files, err := listCabinetFiles(repo, id)
	if err != nil {
		return fmt.Errorf("dump: list: %s", err)
	}

	for _, file := range files {
		fmt.Println(file)
	}
	return nil
}

func dumpCabinet(repo *graph.Repository, id string) error {
	_, data, err := loadCabinet(repo, id)
	if err != nil {
		return fmt.Errorf("dump: %s", err)
	}

	render(stdout, cabDumpTemplate, data)
	return nil
}

var (
	cabDumpTemplate = `Repository:    {{.Repo}}
Created:       {{.Created}} @ {{.Head}}
Dependencies:{{range .Deps}}
  {{.Head}} {{.Pattern}}{{end}}
`
)

func init() {
	cabCmd.Run = cabFunc
}
