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
	"encoding/gob"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"kylelemons.net/go/rx/graph"
)

const rxFileVersion = 1

func defEnv(name, def string) string {
	if os.Getenv(name) == "" {
		return def
	}
	return "$" + name
}

var (
	rescan = flag.Bool("rescan", false, "Force a rescan of repositories")
	rxDir  = flag.String("rxdir", defEnv("RX_DIR", filepath.Join("$HOME", ".rx")), "Directory in which to save state")
	asave  = flag.Bool("autosave", true, "Automatically save dependency graph (disable for concurrent runs)")
	maxAge = flag.Duration("max-age", 1*time.Hour, "Nominal amount of time before a rescan is done")
)

var Deps = graph.New()

func expandRxDir() string {
	return os.ExpandEnv(*rxDir)
}

func Scan() error {
	var (
		stale = time.Since(Deps.LastScan) > *maxAge
		empty = len(Deps.Repository) == 0
		force = *rescan
	)
	if !stale && !empty && !force {
		return nil
	}
	return Deps.Scan("all")
}

func Load() {
	if *rescan {
		return
	}

	graphFile := filepath.Join(expandRxDir(), "graph")

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
	if err := os.MkdirAll(expandRxDir(), 0750); err != nil {
		log.Printf("Save: unable to create rxdir: %s", err)
		return
	}

	// Open the graphFile
	graphFile := filepath.Join(expandRxDir(), "graph")
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
