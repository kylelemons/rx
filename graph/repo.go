package graph

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"kylelemons.net/go/rx/vcs"
)

// A Repository is a version-controlled directory containing one or more packages.
type Repository struct {
	Root     string   // directory containing the repository
	VCS      string   // version control system
	Packages []string // packages contained in this repository
}

// String returns the import pattern matching all packages
func (r *Repository) String() string {
	switch len(r.Packages) {
	case 0:
		return ""
	case 1:
		return r.Packages[0]
	}
	prefix, first := len(r.Packages[0]), r.Packages[0]
	for _, imp := range r.Packages {
		if l := len(imp); l < prefix {
			prefix = l
		}
	}
	for _, imp := range r.Packages[1:] {
		for i := 0; i < prefix; i++ {
			if imp[i] != first[i] {
				prefix = i
				break
			}
		}
	}
	return first[:prefix] + "..."
}

func (r *Repository) Head() (string, error) {
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return "", fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	cmd := exec.Command(tool.Command, tool.Current...)
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("repo: head: %s", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Repository) ToRev(rev string) error {
	// TODO(kevlar): This is getting repetitive...
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	cmd := exec.Command(tool.Command)
	cmd.Dir = r.Root
	for _, arg := range tool.ToRev {
		cmd.Args = append(cmd.Args, tsub(arg, rev))
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("repo: to rev %q: %s", rev, err)
	}
	return nil
}

func (r *Repository) Tags() (TagList, error) {
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return nil, fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	up, err := r.revTags(tool, tool.HeadRev, tool.Updates, tool.UpdatesRegex)
	if err != nil {
		return nil, err
	}
	down, err := r.revTags(tool, tool.HeadRev, tool.TagList, tool.TagListRegex)
	if err != nil {
		return nil, err
	}
	return append(up, down...), nil
}

func (r *Repository) Upgrades() (TagList, error) {
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return nil, fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	return r.revTags(tool, tool.HeadRev, tool.Updates, tool.UpdatesRegex)
}

func (r *Repository) Downgrades() (TagList, error) {
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return nil, fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	return r.revTags(tool, tool.HeadRev, tool.TagList, tool.TagListRegex)
}

func (r *Repository) revTags(tool *vcs.Tool, rev string, command []string, regex string) (TagList, error) {
	cmd := exec.Command(tool.Command)
	cmd.Dir = r.Root
	for _, arg := range command {
		cmd.Args = append(cmd.Args, tsub(arg, rev))
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("repo: list tags: %s", err)
	}
	reg := regexp.MustCompile(regex)
	word := regexp.MustCompile(`[^, ]+`)
	var tags TagList
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) == 0 {
			continue
		}
		match := reg.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		for _, tagName := range word.FindAllString(match[2], -1) {
			tags = append(tags, Tag{
				Name: tagName,
				Rev:  match[1],
			})
		}
	}
	return tags, nil
}

// Package is a subset of cmd/go.Package
// Parsed from `go list`
type Package struct {
	// General
	Dir        string // directory containing package sources
	ImportPath string // import path of package in dir
	Name       string // package name
	Target     string // install path
	Goroot     bool   // is this package in the Go root?
	Standard   bool   // is this package part of the standard Go library?
	Root       string // Go root or Go path dir containing this package
	Incomplete bool   // this package or a dependency has an error

	// Package files
	GoFiles      []string // .go files
	TestGoFiles  []string // _test.go files
	XTestGoFiles []string // currently ignored

	// Package imports
	Imports      []string // import paths used by this package
	TestImports  []string // import paths used by _test.go files in this package
	XTestImports []string // currently ignored

	// Rx specific
	RepoRoot string
}

// Keep returns true if the package should be processed by rx.  Packages are
// not processed if they are in GOROOT, a part of the standard library, or had
// an error when processing them.
func (p *Package) Keep() bool {
	return !p.Goroot && !p.Standard && !p.Incomplete
}

// IsBinary returns true if the package is a binary.
// A package is a binary iff the name of the package is "main".
func (p *Package) IsBinary() bool {
	return p.Name == "main"
}

// IsTestable returns true if the package is testable.
// A package is testable iff there are one or more _test.go files.
func (p *Package) IsTestable() bool {
	// TODO(kevlar): XTestGoFiles ?
	return len(p.TestGoFiles) > 0
}

// DetectVCS attempts to detect which version control system is hosting the
// package, or "" if none is found.  If two are detected, the one with the
// longer root path is chosen.  If more than one have identical length paths,
// the result is undefined.
func (p *Package) DetectVCS() (vcsFound, root string) {
	for name, tool := range vcs.Known {
		cmd := exec.Command(tool.Command, tool.RootDir...)
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

func tsub(tpl string, data interface{}) string {
	b := new(bytes.Buffer)
	t := template.New("help")
	if err := template.Must(t.Parse(tpl)).Execute(b, data); err != nil {
		panic(err)
	}
	return b.String()
}
