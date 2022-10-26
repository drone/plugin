// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import "strings"

// Lookup returns the repository and commit associated
// with the named step and version.
func Lookup(name, version string) (repo string, commit string, ok bool) {
	// find the named step
	plugin_, ok := index[name]
	if !ok {
		return
	}
	// use head commit if no version is provided.
	if version == "" {
		return plugin_.repo, "", ok
	}

	// TODO(bradrydzeski) we should be able to use smart
	// matching based on semantic versioning. For example,
	// if the user specifies version '1' we should be able
	// to match to the latest 1.x.x release.

	version_, ok := plugin_.versions[version]
	if !ok {
		return
	}
	return plugin_.repo, version_, ok
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
	repo     string
	versions map[string]string
}
