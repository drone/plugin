// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

// spec defines the bitrise plugin.
type spec struct {
	Deps struct {
		Brew []struct {
			Name string
		}
		Aptget []struct {
			Name string
		}
	}
	Toolkit struct {
		Bash struct {
			Entryfile string `yaml:"entry_file"`
		}
		Go struct {
			Module      string
			PackageName string `yaml:"package_name"`
		}
	}
	Outputs []map[string]interface{} `yaml:"outputs"`
}
