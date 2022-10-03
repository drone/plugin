// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

// spec defines the bitrise plugin.
type spec struct {
	Deps struct {
		Brew []string
		Apt  []string
	}
	Run struct {
		Docker struct {
			Image string
		}
		Bash struct {
			Path string
			Args []string
		}
		Pwsh struct {
			Path string
			Args []string
		}
		Go struct {
			Module string
		}
	}
}
