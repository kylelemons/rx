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
	"fmt"
	"strings"
	"time"
)

// Graph stores the dependency graph of packages.
type Graph struct {
	// If "a" imports "b", DependsOn["a"]["b"] == true
	DependsOn map[string]map[string]bool

	// If "a" imports "b", UsedBy["b"]["a"] == true
	UsedBy map[string]map[string]bool

	// Package["import/path"] = &Package{...}
	Package map[string]*Package

	// Repository["repo/path"] = &Repository{...}
	Repository map[string]*Repository

	// When this graph was last updated.
	LastScan time.Time
}

func New() *Graph {
	return &Graph{
		DependsOn:  make(map[string]map[string]bool),
		UsedBy:     make(map[string]map[string]bool),
		Package:    make(map[string]*Package),
		Repository: make(map[string]*Repository),
	}
}

// FindRepo attempts to find a repository with the given key.
// If a unique repository is found it is returned, an error otherwise.
func (g *Graph) FindRepo(key string) (*Repository, error) {
	var found *Repository
	// Try suffix match first
	for path, repo := range g.Repository {
		if strings.HasSuffix(path, key) {
			if found != nil {
				return nil, fmt.Errorf("non-unique repository specifier %q", key)
			}
			found = repo
		}
	}
	if found != nil {
		return found, nil
	}

	// ... then generic substring match
	for path, repo := range g.Repository {
		if strings.Contains(path, key) {
			if found != nil {
				return nil, fmt.Errorf("non-unique repository specifier %q", key)
			}
			found = repo
		}
	}
	if found == nil {
		return nil, fmt.Errorf("unknown repository %q", key)
	}
	return found, nil
}

func (g *Graph) traceDeps(repo *Repository, through map[string]map[string]bool) ([]*Repository, error) {
	roots := map[string]bool{}
	for _, ipath := range repo.Packages {
		for dep := range through[ipath] {
			if pkg, ok := g.Package[dep]; ok && pkg.RepoRoot != repo.Root {
				roots[pkg.RepoRoot] = true
			}
		}
	}
	repos := make([]*Repository, 0, len(roots))
	for root := range roots {
		repo, ok := g.Repository[root]
		if !ok {
			return nil, fmt.Errorf("unable to find repository %q", root)
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// RepoDeps returns a list of the repositories which contain packages
// upon which packages in the given repository depend.
func (g *Graph) RepoDeps(repo *Repository) ([]*Repository, error) {
	return g.traceDeps(repo, g.DependsOn)
}

// RepoUsers returns a list of the repositories which contain packages
// which depend on packages in the given repository.
func (g *Graph) RepoUsers(repo *Repository) ([]*Repository, error) {
	return g.traceDeps(repo, g.UsedBy)
}

// addImport adds both directions of an import relationship to the graph.
func (g *Graph) addImport(importer, importee string) {
	if g.DependsOn[importer] == nil {
		g.DependsOn[importer] = make(map[string]bool)
	}
	g.DependsOn[importer][importee] = true
	if g.UsedBy[importee] == nil {
		g.UsedBy[importee] = make(map[string]bool)
	}
	g.UsedBy[importee][importer] = true
}

// addPackage adds a package to the graph and links it up automatically.
func (g *Graph) addPackage(pkg *Package) {
	g.Package[pkg.ImportPath] = pkg

	// Add package to repository
	rep := g.Repository[pkg.RepoRoot]
	rep.Packages = append(rep.Packages, pkg.ImportPath)

	// TODO(kevlar): XTestImports? These can introduce circular deps...
	for _, depList := range [][]string{pkg.Imports, pkg.TestImports} {
		for _, dep := range depList {
			g.addImport(pkg.ImportPath, dep)
		}
	}
}

// delPackage removes references to a Package from the graph.
// It should only be used when also removing the repository
// or when the package will be re-added.
func (g *Graph) delPackage(importPath string) {
	delete(g.Package, importPath)
}

// addRepository adds a Repository to the graph.
// Call this before adding its packages to the graph.
func (g *Graph) addRepository(root, vcs string) {
	if _, ok := g.Repository[root]; ok {
		return
	}
	g.Repository[root] = &Repository{
		Root: root,
		VCS:  vcs,
	}
}

// delRepository removes references to a Repository from the graph.
// The packages from the repository are automatically removed.
func (g *Graph) delRepository(root string) {
	rep, ok := g.Repository[root]
	if !ok {
		return
	}
	for _, importPath := range rep.Packages {
		g.delPackage(importPath)
	}
	delete(g.Repository, root)
}
