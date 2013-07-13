package main

import ()

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
	cpointNum   = cpointCmd.Flag.Int("n", 15, "number of checkpoints to list (0 for all)")
	cpointSave  = cpointCmd.Flag.String("save", "", "save a new checkpoint with the given comment")
	cpointList  = cpointCmd.Flag.Bool("list", false, "list checkpoints")
	cpointApply = cpointCmd.Flag.Int("apply", 0, "apply the specified checkpoint")
)

func cpointFunc(cmd *Command, args ...string) {
	if len(args) > 1 {
		cmd.BadArgs("takes no arguments")
	}

	var err error
	switch {
	case *cpointSave != "":
		cmd.BadArgs("--save unimplemented")
	case *cpointApply != 0:
		cmd.BadArgs("--apply unimplemented")
	case *cpointList:
		cmd.BadArgs("--list unimplemented")
	default:
		cmd.BadArgs("no mode specified")
	}
	if err != nil {
		cmd.Fatalf("%s", err)
	}
}

func init() {
	cpointCmd.Run = cpointFunc
}
