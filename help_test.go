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
	"bytes"
	"regexp"
	"testing"
)

func TestHelp(t *testing.T) {
	tests := []struct {
		Desc  string
		Args  []string
		Regex []string
	}{
		{
			Desc: "Usage",
			Regex: []string{
				`^rx is a command`,
				`\shelp.*Help on the rx`,
			},
		},
		{
			Desc: "No Command",
			Args: []string{"frobber"},
			Regex: []string{
				`error: unknown.*frobber`,
			},
		},
		{
			Desc: "Help help",
			Args: []string{"help"},
			Regex: []string{
				`[options] [command]`,
				`rx.*subcommand`,
				`--godoc`,
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
