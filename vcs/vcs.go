package vcs

type Tool struct {
	// General tool-specific settings
	Command string
	HeadRev string

	// This command lists the repository root and should fail outside a repository
	RootDir []string

	// This command and regex are used to parse commit IDs and tags
	// The command should produce commits in reverse chronological order.
	// Only ancestors of the given revision should be listed.
	// The regex should leave the commit ID in $1 and a comma/whitespace
	// separated list of tags in $2.
	TagList      []string // {{.}} == revision
	TagListRegex string
}

var Known = map[string]*Tool{
	"git": {
		Command: "git",
		HeadRev: "HEAD",
		// Commands
		RootDir: []string{"rev-parse", "--show-toplevel"},
		TagList: []string{"log", "--pretty=format:%H%d", "{{.}}"},
		// Regexes
		TagListRegex: `^([a-z0-9]+) \((.*)\)`,
	},
	"hg": {
		Command: "hg",
		HeadRev: ".",
		// Commands
		RootDir: []string{"root"},
		TagList: []string{"log", "--rev=ancestors({{.}}) and tag()", "--template={node} {tags}\n"},
		// Regexes
		TagListRegex: `^([a-z0-9]+) (.*)`,
	},
}
