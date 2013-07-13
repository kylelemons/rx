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
		cmd.Fatalf("--open: unimplemented")
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
	Repo    string           // The import path prefix covered by this cabinet (usually import/path/...)
	Created time.Time        // The time the cabinet was created
	Head    string           // The hash of the repository at which this cabinet was created
	Deps    []*CabDependency // The dependencies of this package
}

// A CabDependency stores information about a dependency of a repository.
type CabDependency struct {
	Pattern  string   // The pattern required to scan for updates
	Packages []string // Try `go get -d` on these in order until one succeeds
	Head     string   // The hash of the repository to use after installation
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
	deps, err := Deps.RepoDeps(repo)
	if err != nil {
		return fmt.Errorf("build: scan dependencies: %s", err)
	}
	for _, dep := range deps {
		head, err := dep.Head()
		if err != nil {
			return fmt.Errorf("build: get %s head: %s", err)
		}
		data.Deps = append(data.Deps, &CabDependency{
			Pattern:  dep.String(),
			Packages: dep.Packages,
			Head:     head,
		})
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
	files, err := listCabinetFiles(repo, id)
	if err != nil {
		return fmt.Errorf("dump: list: %s", err)
	}
	switch cnt := len(files); {
	case cnt == 0:
		return fmt.Errorf("dump: no matching cabinet files found")
	case cnt > 1:
		return fmt.Errorf("dump: non-unique id pattern %q (matched %d cabinets)", id, cnt)
	}
	filename := files[0]

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("dump: open cabinet: %s", err)
	}

	data := new(CabFile)
	if err := gob.NewDecoder(file).Decode(data); err != nil {
		return fmt.Errorf("dump: decode cabinet: %s", err)
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
