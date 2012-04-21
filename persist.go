package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylelemons/rx/repo"
)

var (
	rescan = flag.Bool("rescan", false, "Force a rescan of repositories")
	rxDir  = flag.String("rxdir", filepath.Join(os.Getenv("HOME"), ".rx"), "Directory in which to save state")
)

var Repos repo.RepoMap

func Scan() error {
	if Repos != nil && !*rescan {
		return nil
	}
	r, err := repo.Scan()
	Repos = r
	return err
}

// TODO(kevlar): environment variable RX_DIR or something

func Load() {
	repoFile := filepath.Join(*rxDir, "repos")
	repo, err := os.Open(repoFile)
	if err != nil {
		return
	}
	if err := gob.NewDecoder(repo).Decode(&Repos); err != nil {
		fmt.Fprintf(stdout, "rx: error loading repos: %s", err)
		os.Remove(repoFile)
	}
}

func Save() {
	if Repos == nil {
		return
	}

	if err := os.MkdirAll(*rxDir, 0750); err != nil {
		fmt.Fprintf(stdout, "rx: unable to create rxdir: %s", err)
		os.Exit(1)
	}

	repoFile := filepath.Join(*rxDir, "repos")
	repo, err := os.Create(repoFile)
	if err != nil {
		fmt.Fprintf(stdout, "rx: error opening .rx/repo file: %s", err)
		os.Exit(1)
	}
	if err := gob.NewEncoder(repo).Encode(Repos); err != nil {
		fmt.Fprintf(stdout, "rx: error encoding repos: %s", err)
		os.Exit(1)
	}
}
