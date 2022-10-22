// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"encoding/json"
	"strings"

	"github.com/drone/plugin/pkg/encoder"
	"github.com/pkg/errors"
)

func GetWith(envVars map[string]string) (map[string]string, error) {
	if val, ok := envVars["PLUGIN_WITH"]; ok {
		with, err := strToMap(val)
		if err != nil {
			return nil, errors.Wrap(err, "with attribute is not of map type with key & value as string")
		}

		return with, nil
	}
	return nil, nil
}

func GetEnv(envVars map[string]string) map[string]string {
	dst := make(map[string]string)
	// convert drone environment variables to bitrise
	// environment variables
	for key, val := range envVars {
		// drone prefixes plugin input parameters, bitrise
		// does not. trim the prefix and convert to lowercase
		// for bitrise compatibility.
		if !strings.HasPrefix(key, "PLUGIN_") {
			dst[key] = val
		}
	}
	return dst
}

func strToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	if s == "" {
		return m, nil
	}

	if err := json.Unmarshal([]byte(s), &m); err != nil {
		m1 := make(map[string]interface{})
		if e := json.Unmarshal([]byte(s), &m1); e != nil {
			return nil, e
		}

		for k, v := range m1 {
			m[k] = encoder.Encode(v)
		}
	}
	return m, nil
}
