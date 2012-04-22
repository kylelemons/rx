package main

import (
	"encoding/gob"
	"flag"
	"log"
	"os"
	"path/filepath"

	"kylelemons.net/go/rx/graph"
)

const rxFileVersion = 1

var (
	rescan = flag.Bool("rescan", false, "Force a rescan of repositories")
	rxDir  = flag.String("rxdir", filepath.Join(os.Getenv("HOME"), ".rx"), "Directory in which to save state")
	asave  = flag.Bool("autosave", true, "Automatically save dependency graph (disable for concurrent runs)")
)

var Deps = graph.New()

func Scan() error {
	if !*rescan {
		return nil
	}
	return Deps.Scan("all")
}

// TODO(kevlar): environment variable RX_DIR or something

func Load() {
	graphFile := filepath.Join(*rxDir, "graph")

	// Open the graphFile
	repo, err := os.Open(graphFile)
	if err != nil {
		log.Printf("Skipping load: %s", err)
		os.Remove(graphFile)
		return
	}
	d := gob.NewDecoder(repo)

	// Check the version
	var version int
	if err := d.Decode(&version); err != nil {
		log.Printf("Load: bad file version: %s", err)
		os.Remove(graphFile)
		return
	} else if got, want := version, rxFileVersion; got != want {
		log.Printf("Load: file version mismatch: %d, want %d", got, want)
		os.Remove(graphFile)
		return
	}

	// Load the dependency graph
	log.Printf("Loading graph from %q...", graphFile)
	if err := d.Decode(&Deps); err != nil {
		log.Printf("Load: error loading graph: %s", err)
		os.Remove(graphFile)
	}
}

func Save() {
	if !*asave {
		return
	}

	// Make sure the directory exists
	if err := os.MkdirAll(*rxDir, 0750); err != nil {
		log.Printf("Save: unable to create rxdir: %s", err)
		return
	}

	// Open the graphFile
	graphFile := filepath.Join(*rxDir, "graph")
	repo, err := os.Create(graphFile)
	if err != nil {
		log.Printf("Save: error opening .rx/repo file: %s", err)
		return
	}
	e := gob.NewEncoder(repo)

	// Write the file version
	if err := e.Encode(rxFileVersion); err != nil {
		log.Printf("Save: error writing version: %s", err)
		return
	}

	// Save the dependency graph
	log.Printf("Saving graph to %q...", graphFile)
	if err := e.Encode(Deps); err != nil {
		log.Printf("Save: error encoding graph: %s", err)
		return
	}
}
