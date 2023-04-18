// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import "errors"

// spec defines the bitrise plugin.
type (
	AptSource struct {
		Key  string `yaml:"key,omitempty"`
		Data string `yaml:"line,omitempty"`
	}

	Apt struct {
		Packages []string     `yaml:"packages,omitempty"`
		Sources  []*AptSource `yaml:"sources,omitempty"`
	}
)

type spec struct {
	Deps struct {
		Brew  []string
		Apt   Apt
		Choco []string
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
		Binary struct {
			Source string
		}
	}
}

// UnmarshalYAML implements the unmarshal interface.
func (v *Apt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var out1 []string
	var out2 = Apt{}

	if err := unmarshal(&out1); err == nil {
		v.Packages = out1
		return nil
	}

	if err := unmarshal(&out2); err == nil {
		v.Packages = out2.Packages
		v.Sources = out2.Sources
		return nil
	}

	return errors.New("failed to unmarshal apt")
}
