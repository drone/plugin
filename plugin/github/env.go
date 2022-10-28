// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/drone/plugin/plugin/internal/encoder"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func getWith(envVars map[string]string) (map[string]string, error) {
	if val, ok := envVars["PLUGIN_WITH"]; ok {
		with, err := strToMap(val)
		if err != nil {
			return nil, errors.Wrap(err, "with attribute is not of map type with key & value as string")
		}

		return with, nil
	}
	return nil, nil
}

func getEnv(envVars map[string]string) map[string]string {
	dst := make(map[string]string)
	for key, val := range envVars {
		if !strings.HasPrefix(key, "PLUGIN_") {
			dst[key] = val
		}
	}
	return dst
}

// exportEnv outputs the environment variables produced by the action.
// Diff of env is calculated before and after the action execution to determine
// the environment variables to export.
func exportEnv(before, after string) error {
	diff := diffEnv(before, after)
	if len(diff) == 0 {
		return nil
	}
	exportFile := os.Getenv("DRONE_ENV")
	if exportFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		exportFile = filepath.Join(home, ".drone_export.env")
	}
	if err := godotenv.Write(diff, exportFile); err != nil {
		return err
	}
	return nil
}

// diffEnv takes diff of environment variables present in
// after and before file
func diffEnv(before, after string) map[string]string {
	beforeEnv, _ := godotenv.Read(before)
	afterEnv, _ := godotenv.Read(after)

	diff := make(map[string]string)
	for k, v := range afterEnv {
		if strings.HasPrefix(k, "GITHUB_") {
			continue
		}

		if a, ok := beforeEnv[k]; ok {
			if v != a {
				diff[k] = v
			}
		} else {
			diff[k] = v
		}
	}
	return diff
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
