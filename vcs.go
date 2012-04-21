package main

type VCS struct {
	Command string
	RootCmd []string
}

var KnownVCS = map[string]*VCS{
	"git": &VCS{
		Command: "git",
		RootCmd: []string{"rev-parse", "--show-toplevel"},
	},
	"hg": &VCS{
		Command: "hg",
		RootCmd: []string{"root"},
	},
}
