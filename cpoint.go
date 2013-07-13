package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	userpkg "os/user"
)

var cpointCmd = &Command{
	Name:    "checkpoint",
	Abbrev:  "cpoint",
	Summary: "Save, list, or restore global repository version snapshots.",
	Help: `The checkpoint command is similar to the cabinet command, except
that it has global scope and does not run tests when saving or applying.

Checkpoints are intended as lightweight ways to save state and to share state
among multiple developers sharing an $RX_DIR.  Checkpoints are created with a
comment and a global, sequential ID which can be used to retrieve or delete it.`,
}

// TODO(kevlar): make a CommandSet mechanism that is used both for the top-level
// commands and can be used for recursive subcommands like this, too.

var (
	cpointNum     = cpointCmd.Flag.Int("n", 15, "number of checkpoints to list (0 for all)")
	cpointSave    = cpointCmd.Flag.String("save", "", "save a new checkpoint with the given comment")
	cpointList    = cpointCmd.Flag.Bool("list", false, "list checkpoints")
	cpointApply   = cpointCmd.Flag.Int("apply", 0, "apply the specified checkpoint")
	cpointDelete  = cpointCmd.Flag.Int("delete", 0, "delete the specified checkpoint")
	cpointFilter  = cpointCmd.Flag.String("filter", ".*", "regular expression to filter saved/restored repositories")
	cpointExclude = cpointCmd.Flag.String("exclude", "^$", "regular expression to exclude saved/restored repositories")
)

func cpointFunc(cmd *Command, args ...string) {
	if len(args) > 1 {
		cmd.BadArgs("takes no arguments")
	}

	var data CPointFile

	// Open the checkpoint file
	filename := filepath.Join(expandRxDir(), "checkpoints")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		cmd.Fatalf("open checkpoint file: %s", err)
	}
	defer file.Close()

	// Read the data if possible (EOF is fine, it means there was no data)
	if err := gob.NewDecoder(file).Decode(&data); err != nil && err != io.EOF {
		cmd.Fatalf("decoding checkpoints: %s", err)
	}

	filter, err := regexp.Compile(*cpointFilter)
	if err != nil {
		cmd.BadArgs("--filter: %s", err)
	}

	exclude, err := regexp.Compile(*cpointExclude)
	if err != nil {
		cmd.BadArgs("--exclude: %s", err)
	}

	switch {
	case *cpointSave != "":
		err = data.Save(*cpointSave, filter, exclude)
	case *cpointApply != 0:
		err = data.Apply(*cpointApply, filter, exclude)
	case *cpointDelete != 0:
		err = data.Delete(*cpointDelete)
	case *cpointList:
		data.List(stdout, *cpointNum)
		return
	default:
		cmd.BadArgs("no mode specified")
		return
	}
	if err != nil {
		cmd.Fatalf("%s", err)
	}

	// Write the data back to the file
	if _, err := file.Seek(0, os.SEEK_SET); err != nil {
		cmd.Fatalf("reset file: %s", err)
	}
	if err := gob.NewEncoder(file).Encode(data); err != nil {
		cmd.Fatalf("encode checkpoints: %s", err)
	}

	// Truncate the file down to the size we wrote
	offset, err := file.Seek(0, os.SEEK_CUR)
	if err != nil {
		cmd.Errorf("ftell: %s", err)
	} else if err := file.Truncate(offset); err != nil {
		cmd.Errorf("ftrunc(%d): %s", offset, err)
	}

	log.Printf("Wrote checkpoints to %q", filename)
}

type CPointFile struct {
	LastID      int             // Checkpoint number of last ID
	Checkpoints map[int]*CPoint // Checkpoint data indexed by ID
}

type CPoint struct {
	Comment  string         // Comment for this checkpoint
	User     string         // User who created the checkpoint
	Created  time.Time      // Creation time of the checkpoint
	Versions []*RepoVersion // Versions of repositories at the time of the checkpoint
}

func (f *CPointFile) List(w io.Writer, max int) {
	const DateFormat = "2006/01/02 15:04:05 MST"

	tw := tabify(w)
	defer tw.Flush()

	for id := f.LastID; id > 0 && max > 0; id-- {
		cpoint, ok := f.Checkpoints[id]
		if !ok {
			continue
		}

		fmt.Fprintf(tw, "%d  \t%s  \t%s  \t%d\trepos  \t%s\n",
			id, cpoint.Created.Format(DateFormat), cpoint.User,
			len(cpoint.Versions), cpoint.Comment)
		max--
	}
}

func (f *CPointFile) Save(comment string, filter, exclude *regexp.Regexp) error {
	now := time.Now()

	var versions []*RepoVersion
	for _, repo := range Deps.Repository {
		rv, err := NewRepoVersion(repo)
		if err != nil {
			return fmt.Errorf("save: %s", err)
		}
		if !filter.MatchString(rv.Pattern) || exclude.MatchString(rv.Pattern) {
			continue
		}
		versions = append(versions, rv)
	}

	user, err := userpkg.Current()
	if err != nil {
		user = &userpkg.User{Username: "unknown_user"}
	}
	host, err := os.Hostname()
	if err != nil {
		host = "unknown_host"
	}

	if f.Checkpoints == nil {
		f.Checkpoints = make(map[int]*CPoint)
	}

	f.LastID++
	f.Checkpoints[f.LastID] = &CPoint{
		Comment:  comment,
		User:     user.Username + "@" + host,
		Created:  now,
		Versions: versions,
	}
	log.Printf("Created checkpoint %d with %d repository versions", f.LastID, len(versions))
	return nil
}

func (f *CPointFile) Apply(id int, filter, exclude *regexp.Regexp) error {
	cpoint, ok := f.Checkpoints[id]
	if !ok {
		return fmt.Errorf("checkpoint %d does not exist", id)
	}

	log.Printf("Restoring checkpoint %d: %s", id, cpoint.Comment)
	log.Printf("Checkpoint was created by %s at %s", cpoint.User, cpoint.Created)

	var failed int
	for _, rv := range cpoint.Versions {
		log.Printf("- %s", rv.Pattern)
		if !filter.MatchString(rv.Pattern) || exclude.MatchString(rv.Pattern) {
			log.Printf("  SKIP")
			continue
		}
		if err := rv.Apply(); err != nil {
			log.Printf("  Failed: %s", err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("apply: failed to pin %d versions", failed)
	}
	return nil
}

func (f *CPointFile) Delete(id int) error {
	if _, ok := f.Checkpoints[id]; !ok {
		return fmt.Errorf("checkpoint %d does not exist", id)
	}
	delete(f.Checkpoints, id)
	return nil
}

func init() {
	cpointCmd.Run = cpointFunc
}
