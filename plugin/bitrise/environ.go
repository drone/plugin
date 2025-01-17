// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"fmt"
	"strings"
	"time"

	"github.com/drone/plugin/plugin/internal/environ"
)

// Environ function converts harness or drone environment
// variables to bitrise environment variables.
func Environ(src []string) []string {
	// make a copy of the environment variables
	dst := environ.Map(src)

	// convert drone environment variables to bitrise
	// environment variables
	for key, val := range dst {
		// drone prefixes plugin input parameters, bitrise
		// does not. trim the prefix and convert to lowercase
		// for bitrise compatibility.
		if strings.HasPrefix(key, "PLUGIN_") {
			key = strings.TrimPrefix("PLUGIN_", "")
			key = strings.ToLower(key)
			dst[key] = val
		}
	}

	bitriseDefaults := map[string]string{
		"BITRISEIO_FINISHED_STAGES":      "",
		"BITRISE_DEPLOY_DIR":             "",
		"BITRISE_APP_TITLE":              "",
		"BITRISE_APP_URL":                "",
		"BITRISE_APP_SLUG":               "",
		"BITRISE_PROVISION_URL":          "",
		"BITRISE_CERTIFICATE_URL":        "",
		"BITRISE_CERTIFICATE_PASSPHRASE": "",
	}

	for key := range bitriseDefaults {
		dst[key] = firstMatch(dst, key)
	}

	// bitrise plugins may depend on bitrise environment variables.
	// map drone environment variables, which are already present
	// in the execution envionment, to their bitrise equivalents.
	//
	// bitrise environment variable docs:
	// https://devcenter.bitrise.io/en/references/available-environment-variables.html
	dst = environ.Combine(dst, map[string]string{
		"CI":                                    "true",
		"PR":                                    "false",
		"BITRISE_IO":                            "true",
		"BITRISE_BUILD_STATUS":                  "0",
		"GIT_REPOSITORY_URL":                    dst["DRONE_REMOTE_URL"],
		"GIT_CLONE_COMMIT_HASH":                 dst["DRONE_COMMIT_SHA"],
		"GIT_CLONE_COMMIT_MESSAGE_SUBJECT":      dst["DRONE_COMMIT_MESSAGE"],
		"GIT_CLONE_COMMIT_MESSAGE_BODY":         "",  // NOT SUPPORTED
		"GIT_CLONE_COMMIT_COUNT":                "1", // NOT SUPPORTED
		"GIT_CLONE_COMMIT_AUTHOR_NAME":          dst["DRONE_COMMIT_AUTHOR_NAME"],
		"GIT_CLONE_COMMIT_AUTHOR_EMAIL":         dst["DRONE_COMMIT_AUTHOR_EMAIL"],
		"GIT_CLONE_COMMIT_COMMITER_NAME":        dst["DRONE_COMMIT_AUTHOR_NAME"],
		"GIT_CLONE_COMMIT_COMMITER_EMAIL":       dst["DRONE_COMMIT_AUTHOR_EMAIL"],
		"BITRISEIO_GIT_REPOSITORY_SLUG":         dst["DRONE_REPO"],
		"BITRISEIO_GIT_REPOSITORY_OWNER":        parseOwner(dst["DRONE_REPO"]),
		"BITRISEIO_GIT_BRANCH_DEST":             dst["DRONE_TARGET_BRANCH"],
		"BITRISEIO_PULL_REQUEST_REPOSITORY_URL": dst["DRONE_COMMIT_LINK"],
		"BITRISEIO_PULL_REQUEST_MERGE_BRANCH":   dst["DRONE_SOURCE_BRANCH"],
		"BITRISEIO_PULL_REQUEST_HEAD_BRANCH":    dst["DRONE_TARGET_BRANCH"],
		"BITRISEIO_PIPELINE_ID":                 firstMatch(dst, "HARNESS_PIPELINE_ID", "DRONE_STAGE_NAME"),
		"BITRISEIO_PIPELINE_TITLE":              firstMatch(dst, "HARNESS_PIPELINE_ID", "DRONE_STAGE_NAME"),
		"BITRISE_GIT_BRANCH":                    dst["DRONE_BRANCH"],
		"BITRISE_GIT_TAG":                       dst["DRONE_TAG"],
		"BITRISE_GIT_COMMIT":                    dst["DRONE_COMMIT_SHA"],
		"BITRISE_GIT_MESSAGE":                   dst["DRONE_COMMIT_MESSAGE"],
		"BITRISE_BUILD_NUMBER":                  dst["DRONE_BUILD_NUMBER"],
		"BITRISE_BUILD_SLUG":                    dst["DRONE_BUILD_NUMBER"],
		"BITRISE_BUILD_TRIGGER_TIMESTAMP":       dst["DRONE_BUILD_CREATED"],                              // MISSING IN HARNESS
		"BITRISE_BUILD_URL":                     dst["DRONE_BUILD_LINK"],                                 // MISSING IN HARNESS
		"BITRISE_PULL_REQUEST":                  dst["DRONE_PULL_REQUEST"],                               // MISSING IN HARNESS
		"BITRISE_SOURCE_DIR":                    firstMatch(dst, "DRONE_WORKSPACE", "HARNESS_WORKSPACE"), // MISSING IN HARNESS                                                   // TODO
		"BITRISE_TRIGGERED_WORKFLOW_ID":         dst["HARNESS_PIPELINE_ID"],
		"BITRISE_TRIGGERED_WORKFLOW_TITLE":      dst["HARNESS_PIPELINE_ID"],
	})

	// is pipeline a pull request?
	if dst["DRONE_PULL_REQUEST"] != "" {
		dst["PR"] = "true"
	}

	// if the build creation timestamp is not present,
	// use the current unix timestamp.
	//
	// TODO remove this once Harness supports DRONE_BUILD_CREATED
	if dst["BITRISE_BUILD_TRIGGER_TIMESTAMP"] == "" {
		dst["BITRISE_BUILD_TRIGGER_TIMESTAMP"] = fmt.Sprint(time.Now().Unix())
	}

	// is pipeline in a failing state?
	if dst["DRONE_BUILD_STATUS"] == "failure" { // MISSING IN HARNESS
		dst["BITRISE_BUILD_STATUS"] = "1"
	}

	return environ.Slice(dst)
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

// helper function gets the owner from a repository slug.
func parseOwner(s string) (owner string) {
	if parts := strings.Split(s, "/"); len(parts) == 2 {
		return parts[0]
	}
	return
}
