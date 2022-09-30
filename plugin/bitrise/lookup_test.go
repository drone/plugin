// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import "testing"

func TestLookup(t *testing.T) {
	repo, commit, ok := Lookup("activate-ssh-key", "3.0.2")
	if !ok {
		t.Errorf("Expect found step")
	}
	if got, want := repo, "https://github.com/bitrise-io/steps-activate-ssh-key.git"; got != want {
		t.Errorf("Expect repository %s, got %s", want, got)
	}
	if got, want := commit, "d4d437de5d7de7cdb4e25116c12fd0344a03923e"; got != want {
		t.Errorf("Expect commit %s, got %s", want, got)
	}
}

func TestLookup_Default(t *testing.T) {
	repo, commit, ok := Lookup("activate-ssh-key", "")
	if !ok {
		t.Errorf("Expect found step")
	}
	if got, want := repo, "https://github.com/bitrise-steplib/steps-activate-ssh-key.git"; got != want {
		t.Errorf("Expect repository %s, got %s", want, got)
	}
	if got, want := commit, "9f0fc00b7a2483a283c0d82106d6816638ac7d41"; got != want {
		t.Errorf("Expect commit %s, got %s", want, got)
	}
}
