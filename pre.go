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
	"log"
	"os"
	"os/exec"
	//"strings"

	"kylelemons.net/go/rx/graph"
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
	preBuild    = preCmd.Flag.Bool("build", true, "build all updated packages")
	preLink     = preCmd.Flag.Bool("link", false, "link and install all updated binaries")
	preTest     = preCmd.Flag.Bool("test", true, "test all updated packages")
	preInstall  = preCmd.Flag.Bool("install", true, "install all updated packages")
	preCascade  = preCmd.Flag.Bool("cascade", true, "recursively process depending packages too")
	preFallback = preCmd.Flag.Bool("rollback", true, "automatically roll back failed upgrade")
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
		if *preFallback && fallback != "" {
			cmd.Errorf("errors detected, falling back to %q...", fallback)
			if err := repo.ToRev(fallback); err != nil {
				cmd.Errorf("during fallback: %s", err)
			}
		}
	}()

	if err := repo.ToRev(repoTag); err != nil {
		cmd.Fatalf("failure to change rev to %q: %s", repoTag, err)
	}

	do := func(repo *graph.Repository, subCmd string) error {
		for _, importPath := range repo.Packages {
			pkg, ok := Deps.Package[importPath]
			if !ok {
				cmd.Fatalf("unknown package %q", importPath)
			}
			switch subCmd {
			case "test":
				if !pkg.IsTestable() {
					continue
				}
				// Install dependencies so we don't get complaints
				if *preInstall {
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

	rebuilt := map[*graph.Repository]bool{}
	process := func(repo *graph.Repository) {
		log.Printf("Processing %s", repo)
		if *preBuild {
			log.Printf(" - Build")
			if err := do(repo, "build"); err != nil {
				cmd.Fatalf("build failed: %q broke %q", repoTag, repo)
			}
		}

		if *preTest {
			log.Printf(" - Test")
			if err := do(repo, "test"); err != nil {
				cmd.Fatalf("test failed: %q broke %q: %s", repoTag, repo, err)
			}
		}

		if *preInstall {
			log.Printf(" - Install")
			if err := do(repo, "install"); err != nil {
				cmd.Fatalf("install failed: %q broke %q", repoTag, repo)
			}
		}

		if *preCascade {
			deps, err := Deps.RepoDeps(repo)
			if err != nil {
				cmd.Fatalf("cascade: %s", err)
			}
			for _, dep := range deps {
				// Don't process repos we've already rebuilt
				if _, ok := rebuilt[dep]; ok {
					continue
				}
				log.Printf(" - Cascade: %s", dep)
				rebuilt[dep] = false
			}
		}
	}

	cascade := true
	rebuilt[repo] = false

	for cascade {
		cascade = false
		for next, processed := range rebuilt {
			if !processed {
				cascade = true
				rebuilt[next] = true
				process(next)
			}
		}
	}

	fallback = ""
}

func init() {
	preCmd.Run = preFunc
}
