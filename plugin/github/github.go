// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package harness provides support for executing Github plugins.
package github

import (
	"os"
	"path/filepath"
)

// Is returns true if the root path is a Harness
// plugin repository.
func Is(root string) bool {
	path := filepath.Join(root, "action.yml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	path = filepath.Join(root, "action.yaml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func getYamlFilename(root string) string {
	if _, err := os.Stat(filepath.Join(root, "action.yml")); err == nil {
		return filepath.Join(root, "action.yml")
	}
	if _, err := os.Stat(filepath.Join(root, "action.yaml")); err == nil {
		return filepath.Join(root, "action.yaml")
	}
	return ""
}
