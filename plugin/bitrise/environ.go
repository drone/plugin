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

	// copy drone environment variable values into their
	// bitrise equivalents.
	for bitrise, drone := range mapping {
		dst[bitrise] = dst[drone]
	}

	// hard-coded bitrise constants
	dst["CI"] = "true"
	dst["PR"] = "false"
	dst["BITRISE_IO"] = "true"
	dst["BITRISE_BUILD_STATUS"] = "0"

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

// maps bitrise variables to drone variables.
// https://devcenter.bitrise.io/en/references/available-environment-variables.html
var mapping = map[string]string{
	"GIT_REPOSITORY_URL":                    "DRONE_REMOTE_URL",
	"GIT_CLONE_COMMIT_HASH":                 "DRONE_COMMIT_SHA",
	"GIT_CLONE_COMMIT_MESSAGE_SUBJECT":      "DRONE_COMMIT_MESSAGE",
	"GIT_CLONE_COMMIT_MESSAGE_BODY":         "",  // do we need this?
	"GIT_CLONE_COMMIT_COUNT":                "1", // do we need this?
	"GIT_CLONE_COMMIT_AUTHOR_NAME":          "DRONE_COMMIT_AUTHOR_NAME",
	"GIT_CLONE_COMMIT_AUTHOR_EMAIL":         "DRONE_COMMIT_AUTHOR_EMAIL",
	"GIT_CLONE_COMMIT_COMMITER_NAME":        "DRONE_COMMIT_AUTHOR_NAME",
	"GIT_CLONE_COMMIT_COMMITER_EMAIL":       "DRONE_COMMIT_AUTHOR_EMAIL",
	"BITRISEIO_GIT_REPOSITORY_OWNER":        "", // TODO
	"BITRISEIO_GIT_REPOSITORY_SLUG":         "", // TODO
	"BITRISEIO_GIT_BRANCH_DEST":             "DRONE_TARGET_BRANCH",
	"BITRISEIO_PULL_REQUEST_REPOSITORY_URL": "", // TODO
	"BITRISEIO_PULL_REQUEST_MERGE_BRANCH":   "", // TODO
	"BITRISEIO_PULL_REQUEST_HEAD_BRANCH":    "", // TODO
	"BITRISEIO_PIPELINE_ID":                 "", // TODO
	"BITRISEIO_PIPELINE_TITLE":              "", // TODO
	"BITRISE_GIT_BRANCH":                    "DRONE_BRANCH",
	"BITRISE_GIT_TAG":                       "DRONE_TAG",
	"BITRISE_GIT_COMMIT":                    "DRONE_COMMIT_SHA",
	"BITRISE_GIT_MESSAGE":                   "DRONE_COMMIT_MESSAGE",
	"BITRISE_BUILD_NUMBER":                  "DRONE_BUILD_NUMBER",
	"BITRISE_BUILD_URL":                     "DRONE_BUILD_LINK",
	"BITRISE_BUILD_SLUG":                    "", // TODO
	"BITRISE_BUILD_TRIGGER_TIMESTAMP":       "", // TODO
	"BITRISE_PULL_REQUEST":                  "DRONE_PULL_REQUEST",
	"BITRISE_SOURCE_DIR":                    "", // TODO
	"BITRISE_DEPLOY_DIR":                    "", // TODO
	"BITRISE_TRIGGERED_WORKFLOW_ID":         "", // TODO
	"BITRISE_TRIGGERED_WORKFLOW_TITLE":      "", // TODO
	"BITRISE_APP_TITLE":                     "", // TODO
	"BITRISE_APP_URL":                       "", // TODO
	"BITRISE_APP_SLUG":                      "", // TODO
	"BITRISE_PROVISION_URL":                 "", // TODO
	"BITRISE_CERTIFICATE_URL":               "", // TODO
	"BITRISE_CERTIFICATE_PASSPHRASE":        "", // TODO
}
