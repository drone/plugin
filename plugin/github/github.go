// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package github provides support for executing GitHub Actions.
package github

import (
	"os"
	"path/filepath"
)

// Is returns true if the root path is a GitHub
// action repository.
func Is(root string) bool {
	path := filepath.Join(root, "action.yml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
