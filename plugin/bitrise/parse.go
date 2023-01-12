// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// helper function to parse the bitrise plugin yaml.
func parse(b []byte) (*spec, error) {
	out := new(spec)
	err := yaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}
	// set the default shell entrypoint if no
	// other default entrypoint or module are
	// provided.
	if out.Toolkit.Go.Module == "" &&
		out.Toolkit.Go.PackageName == "" &&
		out.Toolkit.Bash.Entryfile == "" {
		out.Toolkit.Bash.Entryfile = "step.sh"
	}
	return out, nil
}

// helper function to parse the bitrise plugin yaml file.
func parseFile(s string) (*spec, error) {
	raw, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	return parse(raw)
}

// helper function to parse the bitrise plugin yaml string.
func parseString(s string) (*spec, error) {
	return parse([]byte(s))
}
