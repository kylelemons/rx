package vcs

type Tool struct {
	Command string
	RootCmd []string
}

var Known = map[string]*Tool{
	"git": {
		Command: "git",
		RootCmd: []string{"rev-parse", "--show-toplevel"},
	},
	"hg": {
		Command: "hg",
		RootCmd: []string{"root"},
	},
}
