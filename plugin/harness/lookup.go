// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import "strings"

// Lookup returns the repository and commit associated
// with the named step and version.
func Lookup(name, version string) (repo string, ref string, commit string, ok bool) {
	// find the named step
	plugin_, ok := index[name]
	if !ok {
		return
	}
	// use head commit if no version is provided.
	if version == "" {
		return plugin_.repo, "", "", ok
	}

	// TODO(bradrydzeski) we should be able to use smart
	// matching based on semantic versioning. For example,
	// if the user specifies version '1' we should be able
	// to match to the latest 1.x.x release.

	version_, ok := plugin_.versions[version]
	if !ok {
		return
	}
	return plugin_.repo, "", version_, ok
}

// ParseLookup parses the step string and returns the
// associated repository and commit.
func ParseLookup(s string) (repo string, ref string, commit string, ok bool) {
	// if the strings is prefixed with git:: it means the
	// repository url was provided directly.
	if strings.HasPrefix(s, "git:") || strings.HasPrefix(s, "github.com") ||
		strings.HasPrefix(s, "https://github.com") {
		// trim the git:: prefix
		repo = strings.TrimPrefix(s, "git::")

		// extract the version from the string, if provided.
		if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
			repo = parts[0]
			if strings.HasPrefix(parts[1], "refs/") {
				ref = parts[1]
			} else {
				commit = parts[1]
			}
		}

		// prepend the https scheme if not includes in the
		// github repository url.
		if strings.HasPrefix(repo, "github.com") {
			repo = "https://" + repo

			// append the .git suffix to the github
			// repository url if not provided.
			if !strings.HasSuffix(repo, ".git") {
				repo = repo + ".git"
			}
		}

		return repo, ref, commit, true
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
