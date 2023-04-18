// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"io/ioutil"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

// helper function to parse the bitrise plugin yaml.
func parse(b []byte) (*spec, error) {
	out := new(spec)
	err := yaml.Unmarshal(b, out)
	return out, err
}

// helper function to parse the bitrise plugin yaml file.
func parseFile(s string) (*spec, error) {
	raw, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	slog.Debug("parsing plugin.yml", slog.String("file", string(raw)))
	return parse(raw)
}

// helper function to parse the bitrise plugin yaml string.
func parseString(s string) (*spec, error) {
	return parse([]byte(s))
}
