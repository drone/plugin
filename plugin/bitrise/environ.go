// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"strings"

	"github.com/drone/plugin/plugin/internal/environ"
)

// helper function converts harness environment variables
// to bitrise environment variables.
func convertEnv(src []string) []string {
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
		"GIT_CLONE_COMMIT_MESSAGE_BODY":         "", // TODO
		"GIT_CLONE_COMMIT_COUNT":                "1",
		"GIT_CLONE_COMMIT_AUTHOR_NAME":          dst["DRONE_COMMIT_AUTHOR_NAME"],
		"GIT_CLONE_COMMIT_AUTHOR_EMAIL":         dst["DRONE_COMMIT_AUTHOR_EMAIL"],
		"GIT_CLONE_COMMIT_COMMITER_NAME":        dst["DRONE_COMMIT_AUTHOR_NAME"],
		"GIT_CLONE_COMMIT_COMMITER_EMAIL":       dst["DRONE_COMMIT_AUTHOR_EMAIL"],
		"BITRISEIO_GIT_REPOSITORY_OWNER":        "", // TODO
		"BITRISEIO_GIT_REPOSITORY_SLUG":         "", // TODO
		"BITRISEIO_GIT_BRANCH_DEST":             dst["DRONE_TARGET_BRANCH"],
		"BITRISEIO_PULL_REQUEST_REPOSITORY_URL": "", // TODO
		"BITRISEIO_PULL_REQUEST_MERGE_BRANCH":   dst["DRONE_SOURCE_BRANCH"],
		"BITRISEIO_PULL_REQUEST_HEAD_BRANCH":    dst["DRONE_TARGET_BRANCH"],
		"BITRISEIO_PIPELINE_ID":                 "", // TODO
		"BITRISEIO_PIPELINE_TITLE":              "", // TODO
		"BITRISE_GIT_BRANCH":                    dst["DRONE_BRANCH"],
		"BITRISE_GIT_TAG":                       dst["DRONE_TAG"],
		"BITRISE_GIT_COMMIT":                    dst["DRONE_COMMIT_SHA"],
		"BITRISE_GIT_MESSAGE":                   dst["DRONE_COMMIT_MESSAGE"],
		"BITRISE_BUILD_NUMBER":                  dst["DRONE_BUILD_NUMBER"],
		"BITRISE_BUILD_URL":                     dst["DRONE_BUILD_LINK"],
		"BITRISE_BUILD_SLUG":                    "", // TODO
		"BITRISE_BUILD_TRIGGER_TIMESTAMP":       "", // TODO
		"BITRISE_PULL_REQUEST":                  dst["DRONE_PULL_REQUEST"],
		"BITRISE_SOURCE_DIR":                    dst["DRONE_WORKSPACE"],
		"BITRISE_DEPLOY_DIR":                    "", // TODO
		"BITRISE_TRIGGERED_WORKFLOW_ID":         "", // TODO
		"BITRISE_TRIGGERED_WORKFLOW_TITLE":      "", // TODO
		"BITRISE_APP_TITLE":                     "", // TODO
		"BITRISE_APP_URL":                       "", // TODO
		"BITRISE_APP_SLUG":                      "", // TODO
		"BITRISE_PROVISION_URL":                 "", // TODO
		"BITRISE_CERTIFICATE_URL":               "", // TODO
		"BITRISE_CERTIFICATE_PASSPHRASE":        "", // TODO
	})

	// is pipeline a pull request?
	if dst["DRONE_PULL_REQUEST"] != "" {
		dst["PR"] = "true"
	}

	// is pipeline in a failing state?
	if dst["DRONE_BUILD_STATUS"] == "failure" {
		dst["BITRISE_BUILD_STATUS"] = "1"
	}

	return environ.Slice(dst)
}
