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
		ref  string
	}{
		{
			name: "git::https://github.com/ocotcat/hello-world.git",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
			ref:  "",
		},
		{
			name: "github.com/ocotcat/hello-world.git",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
			ref:  "",
		},
		{
			name: "github.com/ocotcat/hello-world",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "",
			ref:  "",
		},
		{
			name: "github.com/ocotcat/hello-world@a8b99ea78c8f99326516d6b875075ead642b4ca5",
			repo: "https://github.com/ocotcat/hello-world.git",
			hash: "a8b99ea78c8f99326516d6b875075ead642b4ca5",
			ref:  "",
		},
		{
			name: "github.com/drone-plugins/drone-s3@refs/heads/master",
			repo: "https://github.com/drone-plugins/drone-s3.git",
			hash: "",
			ref:  "refs/heads/master",
		},
		{
			name: "github.com/drone-plugins/drone-s3@refs/tags/v1",
			repo: "https://github.com/drone-plugins/drone-s3.git",
			hash: "",
			ref:  "refs/tags/v1",
		},
	}
	for _, test := range tests {
		repo, ref, commit, _ := ParseLookup(test.name)
		if got, want := repo, test.repo; got != want {
			t.Errorf("Expect repository %s, got %s", want, got)
		}
		if got, want := commit, test.hash; got != want {
			t.Errorf("Expect commit %s, got %s", want, got)
		}

		if got, want := ref, test.ref; got != want {
			t.Errorf("Expect ref %s, got %s", want, got)
		}
	}
}
