// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import "testing"

func TestParseLookup(t *testing.T) {
	tests := []struct {
		name string
		repo string
		hash string
	}{
		{
			name: "git::https://github.com/ocotcat/hello-world.git",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
		},
		{
			name: "github.com/ocotcat/hello-world.git",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
		},
		{
			name: "github.com/ocotcat/hello-world",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
		},
	}
	for _, test := range tests {
		repo, commit, _ := ParseLookup(test.name)
		if got, want := repo, test.repo; got != want {
			t.Errorf("Expect repository %s, got %s", want, got)
		}
		if got, want := commit, test.hash; got != want {
			t.Errorf("Expect commit %s, got %s", want, got)
		}
	}
}
