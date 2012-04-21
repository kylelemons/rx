package main

import (
	"os"
	"strings"
	"os/exec"

	"github.com/kylelemons/rx/repo"
)

var preCmd = &Command{
	Name:    "prescribe",
	Usage:   "repo tag",
	Summary: "Update the repository to the given tag/rev.",
	Help: `The prescribe command updates the repository to the named tag or
revision.  The [repo] can be the suffix of the repository root path,
as long as it is unique.  The [tag] is anything understood by the
underlying version control system as a commit, usually a tag, branch,
or commit.

After updating, prescribe will test, build, and the install each package
in the updated repository.  These steps can be disabled via flags such as
"rx prescribe --test=false repo tag".  If a step is disabled, the next
steps will be disabled as well.`,
}

var (
	preTest  = preCmd.Flag.Bool("test", true, "test all updated packages")
	preBuild = preCmd.Flag.Bool("build", true, "build all updated packages")
	preInst  = preCmd.Flag.Bool("install", true, "install all updated packages")
)

func preFunc(cmd *Command, args ...string) {
	if len(args) != 2 {
		cmd.BadArgs("requires two arguments")
	}
	pathSuffix := args[0]
	repoTag := args[1]

	// Scan before accessing Repos
	if err := Scan(); err != nil {
		cmd.Fatalf("scan: %s", err)
	}

	// TODO(kevlar): This seems like something we'll be doing often...
	var rep *repo.Repository
	for p, r := range Repos {
		if strings.HasSuffix(p, pathSuffix) {
			if rep != nil {
				cmd.Fatalf("non-unique suffix %q", pathSuffix)
			}
			rep = r
		}
	}
	if rep == nil {
		cmd.Fatalf("unknown repo %q", pathSuffix)
	}

	fallback, err := rep.Head()
	if err != nil {
		cmd.Fatalf("failure to determine head: %s", err)
	}
	// TODO(kevlar): defer a fallback

	if err := rep.ToRev(repoTag); err != nil {
		cmd.Fatalf("failure to change rev to %q: %s", repoTag, err)
	}

	do := func(subCmd string) error {
		for _, pkg := range rep.Packages {
			cmd := exec.Command("go", "build", pkg.ImportPath)
			cmd.Dir = os.TempDir()
			cmd.Stdout = stdout
			cmd.Stderr = stdout
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !*preBuild {
		return
	}
	if err := do("build"); err != nil {
		rep.ToRev(fallback)
		cmd.Fatalf("build failed")
	}

	if !*preTest {
		return
	}
	if err := do("test"); err != nil {
		rep.ToRev(fallback)
		cmd.Fatalf("test failed")
	}

	if !*preInst {
		return
	}
	if err := do("install"); err != nil {
		rep.ToRev(fallback)
		cmd.Fatalf("install failed")
	}
}

func init() {
	preCmd.Run = preFunc
}
