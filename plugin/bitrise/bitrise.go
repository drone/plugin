// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"os"
	"path/filepath"
)

// Is returns true if the root path is a Bitrise
// plugin repository.
func Is(root string) bool {
	path := filepath.Join(root, "step.yml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
