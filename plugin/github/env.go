// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/drone/plugin/plugin/internal/encoder"
	"github.com/drone/plugin/plugin/internal/environ"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

// Environ function converts harness or drone environment
// variables to github action environment variables.
func Environ(src []string) []string {
	// make a copy of the environment variables
	dst := environ.Map(src)

	// convert drone environment variables to github action
	// environment variables
	for key, val := range dst {
		// drone prefixes plugin input parameters, github action
		// does not. trim the prefix and convert to lowercase
		// for github action compatibility.
		if strings.HasPrefix(key, "PLUGIN_") {
			key = strings.TrimPrefix("PLUGIN_", "")
			key = strings.ToLower(key)
			dst[key] = val
		}
	}

	tagName := dst["DRONE_TAG"]
	branchName := dst["DRONE_BRANCH"]

	arch := "X64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}

	ostype := "Linux"
	runner_tool_cache := "/opt/hostedtoolcache"

	if runtime.GOOS == "darwin" {
		ostype = "macOS"
		runner_tool_cache = "/Users/anka/hostedtoolcache"
	} else if runtime.GOOS == "windows" {
		ostype = "Windows"
		runner_tool_cache = "C:\\hostedtoolcache\\windows"
	}

	// github actions may depend on github action environment variables.
	// map drone environment variables, which are already present
	// in the execution envionment, to their github action equivalents.
	//
	// github action environment variable docs:
	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	dst = environ.Combine(dst, map[string]string{
		"GITHUB_BASE_REF":         dst["DRONE_TARGET_BRANCH"],
		"GITHUB_HEAD_REF":         dst["DRONE_SOURCE_BRANCH"],
		"GITHUB_REF":              dst["DRONE_COMMIT_REF"],
		"GITHUB_REPOSITORY":       dst["DRONE_REPO"],
		"GITHUB_REPOSITORY_OWNER": parseOwner(dst["DRONE_REPO"]),
		"SHA_REF":                 dst["DRONE_COMMIT_SHA"],
		"GITHUB_RUN_ATTEMPT":      dst["DRONE_BUILD_NUMBER"],
		"GITHUB_RUN_NUMBER":       dst["DRONE_BUILD_NUMBER"],
		"GITHUB_WORKSPACE":        firstMatch(dst, "DRONE_WORKSPACE", "HARNESS_WORKSPACE"),
		"RUNNER_ARCH":             arch,
		"RUNNER_OS":               ostype,
		"RUNNER_NAME":             "HARNESS HOSTED",
		"RUNNER_TOOL_CACHE":       runner_tool_cache,
	})

	if tagName != "" {
		dst["GITHUB_REF_NAME"] = tagName
		dst["GITHUB_REF_TYPE"] = "tag"
	} else if branchName != "" {
		dst["GITHUB_REF_NAME"] = branchName
		dst["GITHUB_REF_TYPE"] = "branch"
	}

	return environ.Slice(dst)
}

// helper function gets the owner from a repository slug.
func parseOwner(s string) (owner string) {
	if parts := strings.Split(s, "/"); len(parts) == 2 {
		return parts[0]
	}
	return
}

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
	log := slog.Default()

	beforeEnv, err := godotenv.Read(before)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to read before env file: %s", err))
	}
	afterEnv, err := godotenv.Read(after)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to read after env file: %s", err))
	}

	diffB64 := make(map[string]string)
	for k, v := range afterEnv {
		if strings.HasPrefix(k, "GITHUB_") {
			continue
		}

		if a, ok := beforeEnv[k]; ok {
			if v != a {
				diffB64[k] = v
			}
		} else {
			diffB64[k] = v
		}
	}

	encoding := base64.StdEncoding
	if runtime.GOOS == "windows" {
		encoding = base64.RawURLEncoding
	}

	// Base64 decode env values
	diff := make(map[string]string)
	for k, v := range diffB64 {
		data, err := encoding.DecodeString(v)
		if err == nil {
			diff[k] = string(data)
		} else {
			log.Warn(fmt.Sprintf("failed to decode env value: %s", string(data)))
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

// helper function find the first matching environment
// variable in the map.
func firstMatch(envs map[string]string, keys ...string) (val string) {
	for _, key := range keys {
		if env, ok := envs[key]; ok {
			return env
		}
	}
	return
}
