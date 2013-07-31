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
	"time"
)

// Scan scans the named packages and updats their records in the dependency
// graph.  To scan everything, use Scan("all").
func (g *Graph) Scan(target string) error {
	start := time.Now()
	defer func() {
		log.Printf("Scan took %s", time.Since(start))
	}()

	// Set the time first so that no changes to the scanned directories
	// will have happened before LastScan.
	g.LastScan = start

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
