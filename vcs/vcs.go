package vcs

type Tool struct {
	// General tool-specific settings
	Command string
	HeadRev string

	// This command lists the repository root and should fail outside a repository.
	RootDir []string

	// This command updates to the given revision
	ToRev []string // {{.}} == revision

	// This command returns an absolute commit identifier for the current HEAD.
	Current []string

	// This command and regex are used to parse commit IDs and tags.
	// The command should produce commits in reverse chronological order.
	// Only ancestors of the given revision should be listed.
	// The regex should leave the commit ID in $1 and a comma/whitespace
	// separated list of tags in $2.
	TagList      []string // {{.}} == revision
	TagListRegex string

	// This command is identical to TagList except it lists tags
	// for which the given revision is an ancestor.
	Updates      []string // {{.}} == revision
	UpdatesRegex string
}

var Known = map[string]*Tool{
	"git": {
		Command: "git",
		HeadRev: "HEAD",
		// Commands
		RootDir: []string{"rev-parse", "--show-toplevel"},
		ToRev:   []string{"checkout", "{{.}}"},
		Current: []string{"log", "--pretty=format:%H", "HEAD"},
		TagList: []string{"log", "--pretty=format:%H%d", "{{.}}"},
		Updates: []string{"log", "--pretty=format:%H%d", "--all", "^{{.}}"},
		// Regexes
		TagListRegex: `^([a-z0-9]+) \((.*)\)`,
		UpdatesRegex: `^([a-z0-9]+) \((.*)\)`,
	},
	"hg": {
		Command: "hg",
		HeadRev: ".",
		// Commands
		RootDir: []string{"root"},
		ToRev:   []string{"update", "{{.}}"},
		Current: []string{"log", "--template={node}", "--rev=."},
		TagList: []string{"log", "--template={node} {tags}\n", "--rev=reverse(ancestors({{.}}))   and branch({{.}}) and tag()"},
		Updates: []string{"log", "--template={node} {tags}\n", "--rev=reverse(descendants({{.}})) and branch({{.}}) and tag() and not {{.}}"},
		// Regexes
		TagListRegex: `^([a-z0-9]+) (.*)`,
		UpdatesRegex: `^([a-z0-9]+) (.*)`,
	},
}
