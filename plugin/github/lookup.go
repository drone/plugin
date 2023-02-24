// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"net/url"
	"strings"
)

// ParseLookup parses the step string and returns the
// associated repository and ref.
func ParseLookup(s string) (repo string, ref string, ok bool) {
	if !strings.HasPrefix(s, "https://github.com") {
		s, _ = url.JoinPath("https://github.com", s)
	}

	if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return s, "", true
}
