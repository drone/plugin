// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package harness provides support for executing Harness plugins.
package harness

import (
	"os"
	"path/filepath"
)

// Is returns true if the root path is a Harness
// plugin repository.
func Is(root string) bool {
	path := filepath.Join(root, "plugin.yml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
