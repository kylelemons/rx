package main

import (
	"testing"
	"bytes"
	"regexp"
)

func TestHelp(t *testing.T) {
	tests := []struct{
		Desc  string
		Args  []string
		Regex []string
	}{
		{
			Desc: "Usage",
			Regex: []string{
				"^rx is a command",
				"  help.*Help on the rx",
			},
		},
		{
			Desc: "No Command",
			Args: []string{"frobber"},
			Regex: []string{
				"error: unknown.*frobber",
			},
		},
		{
			Desc: "Help help",
			Args: []string{"help"},
			Regex: []string{
				"[options] [command]",
				"rx.*subcommand",
				"--godoc",
			},
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		stdout = buf
		helpCmd.Exec(test.Args)
		out := buf.String()
		for i, r := range test.Regex {
			matched, err := regexp.MatchString(r, out)
			if err != nil {
				t.Errorf("%s: %q: %s", r, err)
			}
			if !matched {
				t.Errorf("%s: regexp[%d] failed: %q\nOutput:\n%s", test.Desc, i, r, out)
			}
		}
	}
}
