package repo

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/kylelemons/rx/vcs"
)

// A Repository is a directory 
type Repository struct {
	Path     string     // directory containing the repository
	VCS      string     // Version control system
	Packages []*Package // packages contained in this repository
	RepoDeps []string   // repositories (by path) that packages
}

func (r *Repository) Head() (string, error) {
	tool, ok := vcs.Known[r.VCS]
	if !ok {
		return "", fmt.Errorf("repo: unknown vcs %q", r.VCS)
	}
	cmd := exec.Command(tool.Command, tool.Current...)
	cmd.Dir = r.Path
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
	cmd.Dir = r.Path
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
	cmd.Dir = r.Path
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
