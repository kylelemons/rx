package main

import (
	//"log"
	//"os"
	//"os/exec"
	//"strings"

	//"kylelemons.net/go/rx/graph"
)

var preCmd = &Command{
	Name:    "prescribe",
	Usage:   "<repo> <tag>",
	Summary: "Update the repository to the given tag/rev.",
	Help: `The prescribe command updates the repository to the named tag or
revision.  The <repo> can be any piece of the repository root path, as long as
it is unique.  The <tag> is anything understood by the underlying version
control system as a commit, usually a tag, branch, or commit.

After updating, prescribe will test, build, and the install each package
in the updated repository.  These steps can be disabled via flags such as
"rx prescribe --test=false repo tag".

By default, this will not link and install affected binaries; to turn this
behavior on, see the --link option.`,
}

var (
	preBuild = preCmd.Flag.Bool("build", true, "build all updated packages")
	preLink  = preCmd.Flag.Bool("link", false, "link and install all updated binaries")
	preTest  = preCmd.Flag.Bool("test", true, "test all updated packages")
	preInst  = preCmd.Flag.Bool("install", true, "install all updated packages")
	preCasc  = preCmd.Flag.Bool("cascade", true, "recursively process depending packages too")
)

func preFunc(cmd *Command, args ...string) {
	if len(args) != 2 {
		cmd.BadArgs("requires two arguments")
	}
	path := args[0]
	repoTag := args[1]

	repo, err := Deps.FindRepo(path)
	if err != nil {
		cmd.Fatalf("<repo>: %s", err)
	}

	fallback, err := repo.Head()
	if err != nil {
		cmd.Fatalf("failure to determine head: %s", err)
	}
	defer func() {
		if fallback != "" {
			cmd.Errorf("errors detected, falling back to %q...", fallback)
			if err := repo.ToRev(fallback); err != nil {
				cmd.Errorf("during fallback: %s", err)
			}
		}
	}()

	if err := repo.ToRev(repoTag); err != nil {
		cmd.Fatalf("failure to change rev to %q: %s", repoTag, err)
	}

	/*
	do := func(r *graph.Repository, subCmd string) error {
		for _, pkg := range r.Packages {
			switch subCmd {
			case "test":
				if !pkg.IsTestable() {
					continue
				}
				// Install dependencies so we don't get complaints
				if *preInst {
					exec.Command("go", "test", "-i", pkg.ImportPath).Run()
				}
			case "install":
				if !*preLink && pkg.IsBinary() {
					continue
				}
			}
			log.Printf("   - %s", pkg.ImportPath)
			cmd := exec.Command("go", subCmd, pkg.ImportPath)
			cmd.Dir = os.TempDir()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		return nil
	}

	rebuilt := map[*repo.Graphitory]bool{}
	process := func(r *repo.Graphitory) {
		log.Printf("Processing repo in %q...", r.Path)
		if *preBuild {
			log.Printf(" - Build")
			if err := do(r, "build"); err != nil {
				cmd.Fatalf("build failed: %q on %q broke %q",
					repoTag, rep.Path, r.Path)
			}
		}

		if *preTest {
			log.Printf(" - Test")
			if err := do(r, "test"); err != nil {
				cmd.Fatalf("test failed: %q on %q broke %q: %s",
					repoTag, rep.Path, r.Path, err)
			}
		}

		if *preInst {
			log.Printf(" - Install")
			if err := do(r, "install"); err != nil {
				cmd.Fatalf("install failed: %q on %q broke %q",
					repoTag, rep.Path, r.Path)
			}
		}

		if *preCasc {
			log.Printf(" - Cascade")
			for _, check := range Graph {
				// Don't check repos we've already rebuilt
				if _, ok := rebuilt[check]; ok {
					continue
				}
				for _, dep := range check.RepoDeps {
					// If repo `check` depends on the current repo `r`
					if dep == r.Path {
						// rebuild it
						log.Printf("   - Graphitory %q", check.Path)
						rebuilt[check] = false
						break
					}
				}
			}
		}
	}

	cascade := true
	rebuilt[rep] = false

	for cascade {
		cascade = false
		for rep, processed := range rebuilt {
			if !processed {
				cascade = true
				rebuilt[rep] = true
				process(rep)
			}
		}
	}
	*/

	fallback = ""
}

func init() {
	preCmd.Run = preFunc
}
