// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import "testing"

func TestIs(t *testing.T) {
	tests := map[string]bool{
		"bar":                                false,
		"foo/bar":                            false,
		"company.com/foo/bar":                false,
		"git://docker.io/foo/bar":            false,
		"git+http://docker.io/foo/bar":       false,
		"git+https://docker.io/foo/bar":      false,
		"docker://company.com/foo/bar":       true,
		"docker+http://company.com/foo/bar":  true,
		"docker+https://company.com/foo/bar": true,
		"docker.io/foo/bar":                  true,
		"http://docker.io/foo/bar":           true,
		"https://docker.io/foo/bar":          true,
		"quay.io/foo/bar":                    true,
		"gcr.io/foo/bar":                     true,
		"us.gcr.io/foo/bar":                  true,
		"eu.gcr.io/foo/bar":                  true,
		"asia.gcr.io/foo/bar":                true,
		"azurecr.io/foo/bar":                 true,
		"public.ecr.aws/foo/bar":             true,
		"*.dkr.ecr.*.amazonaws.com/**":       true,
	}
	for k, v := range tests {
		got, want := Is(k), v
		if got != want {
			t.Errorf("Want %q is image %v, got %v", k, want, got)
		}
	}
}
