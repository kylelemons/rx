package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"sync"
)

// Scan scans the named packages and updats their records in the dependency
// graph.  To scan everything, use Scan("all").
func (g *Graph) Scan(target string) error {
	list := exec.Command("go", "list", "-e", "-json", target)
	list.Stderr = os.Stderr
	js, err := list.Output()
	if err != nil {
		return fmt.Errorf("repo: go list %q: %s", target, err)
	}
	dec := json.NewDecoder(bytes.NewReader(js))

	// A type for the completed processing
	type found struct {
		*Package
		vcs string
	}

	// Make a channel to send the completed packages on
	pkgs := make(chan found, 32)

	// Sync up on the completed processing
	wg := sync.WaitGroup{}

	process := func(pkg *Package) {
		defer wg.Done()

		// Only save packages we want to keep
		if !pkg.Keep() {
			log.Printf("Skipping %q", pkg.ImportPath)
			return
		}
		log.Printf("Adding %q", pkg.ImportPath)

		// Detect the version control system
		vcs, root := pkg.DetectVCS()
		pkg.RepoRoot = root

		// Ignore things we don't understand :D
		if vcs == "" {
			return
		}

		pkgs <- found{pkg, vcs}
	}

	for {
		pkg := new(Package)

		// Decode the next package
		err := dec.Decode(pkg)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("repo: error parsing package: %s", err)
			break
		}

		wg.Add(1)
		go process(pkg)
	}

	// Wait on repo detection, then close
	go func() {
		wg.Wait()
		close(pkgs)
	}()

	seen := map[string]bool{}
	for f := range pkgs {
		if !seen[f.RepoRoot] {
			g.delRepository(f.RepoRoot)
			g.addRepository(f.RepoRoot, f.vcs)
		}
		seen[f.RepoRoot] = true
		g.addPackage(f.Package)
	}

	for root := range seen {
		sort.Strings(g.Repository[root].Packages)
	}
	return nil
}
