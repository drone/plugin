// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import "strings"

//go:generate go run ../../scripts/bitrise.go

// Lookup returns the repository and commit associated
// with the named step and version.
func Lookup(name, version string) (repo string, commit string, ok bool) {
	// find the named step
	plugin_, ok := index[name]
	if !ok {
		return
	}
	// use default version if none provided
	if version == "" {
		version = plugin_.version
	}
	release_, ok := plugin_.releases[version]
	if !ok {
		return
	}
	return release_.repo, release_.commit, ok
}

// ParseLookup parses the step string and returns the
// associated repository and commit.
func ParseLookup(s string) (repo string, commit string, ok bool) {
	// if the strings is prefixed with git:: it means the
	// repository url was provided directly.
	if strings.HasPrefix(s, "git::") {
		// trim the git:: prefix
		s = strings.TrimPrefix(s, "git::")

		// extract the version from the string, if provided.
		if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
			return parts[0], parts[1], true
		}
		return s, "", true
	}

	if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
		return Lookup(parts[0], parts[1])
	}
	return Lookup(s, "")
}

type plugin struct {
	name     string
	version  string
	releases map[string]release
}

type release struct {
	version string
	repo    string
	commit  string
}
