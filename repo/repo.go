package repo

import (
	"os/exec"
	"strings"

	"github.com/kylelemons/rx/vcs"
)

// A Repository is a directory 
type Repository struct {
	Path     string     // directory containing the repository
	VCS      string     // Version control system
	Packages []*Package // packages contained in this repository
	RepoDeps []string   // repositories (by path) that packages
}

// Package is a subset of cmd/go.Package
type Package struct {
	// Parsed from `go list`
	Dir        string   // directory containing package sources
	ImportPath string   // import path of package in dir
	Name       string   // package name
	Target     string   // install path
	Goroot     bool     // is this package in the Go root?
	Standard   bool     // is this package part of the standard Go library?
	Root       string   // Go root or Go path dir containing this package
	Imports    []string // import paths used by this package
	Incomplete bool     // this package or a dependency has an error
}

// Keep returns true if the package should be processed by rx.  Packages are
// not processed if they are in GOROOT, a part of the standard library, is the
// "main" package, or had an error when processing them.
func (p *Package) Keep() bool {
	return !p.Goroot && !p.Standard && !p.Incomplete && p.Name != "main"
}

// DetectVCS attempts to detect which version control system is hosting the
// package, or "" if none is found.  If two are detected, the one with the
// longer root path is chosen.  If more than one have identical length paths,
// the result is undefined.
func (p *Package) DetectVCS() (vcsFound, root string) {
	for name, tool := range vcs.Known {
		cmd := exec.Command(tool.Command, tool.RootCmd...)
		cmd.Dir = p.Dir
		b, err := cmd.Output()
		if err != nil {
			continue
		}
		if len(b) > len(root) {
			vcsFound, root = name, strings.TrimSpace(string(b))
		}
	}
	return vcsFound, root
}
