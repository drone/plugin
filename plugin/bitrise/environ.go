// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import "strings"

// helper function converts harness environment variables
// to bitrise environment variables.
func convertEnv(src []string) []string {
	// make a copy of the environment variables
	dst := src[:]

	// convert drone environment variables to bitrise
	// environment variables
	for _, env := range src {
		// separate the key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		// drone prefixes plugin input parameters, bitrise
		// does not. trim the prefix and convert to lowercase
		// for bitrise compatibility.
		if strings.HasPrefix(key, "PLUGIN_") {
			key = strings.TrimPrefix("PLUGIN_", "")
			key = strings.ToLower(key)
			dst = append(dst, key+"="+val)
		}
	}

	return dst
}

// BITRISE_TRIGGERED_WORKFLOW_ID
// BITRISE_TRIGGERED_WORKFLOW_TITLE
// BITRISE_BUILD_STATUS
// BITRISE_SOURCE_DIR
// BITRISE_DEPLOY_DIR
// CI
// PR

// BITRISE_BUILD_NUMBER
// BITRISE_APP_TITLE
// BITRISE_APP_URL
// BITRISE_APP_SLUG
// BITRISE_BUILD_URL
// BITRISE_BUILD_SLUG
// BITRISE_BUILD_TRIGGER_TIMESTAMP
// GIT_REPOSITORY_URL
// BITRISE_GIT_BRANCH
// BITRISEIO_GIT_BRANCH_DEST
// BITRISE_GIT_TAG
// BITRISE_GIT_COMMIT
// BITRISE_GIT_MESSAGE
// BITRISEIO_GIT_REPOSITORY_OWNER
// BITRISEIO_GIT_REPOSITORY_SLUG
// BITRISE_PULL_REQUEST
// BITRISEIO_PULL_REQUEST_REPOSITORY_URL
// BITRISEIO_PULL_REQUEST_MERGE_BRANCH
// BITRISEIO_PULL_REQUEST_HEAD_BRANCH
// BITRISE_PROVISION_URL
// BITRISE_CERTIFICATE_URL
// BITRISE_CERTIFICATE_PASSPHRASE
// BITRISE_IO
// BITRISEIO_PIPELINE_ID
// BITRISEIO_PIPELINE_TITLE

// GIT_CLONE_COMMIT_HASH
// GIT_CLONE_COMMIT_MESSAGE_SUBJECT
// GIT_CLONE_COMMIT_MESSAGE_BODY
// GIT_CLONE_COMMIT_COUNT
// GIT_CLONE_COMMIT_AUTHOR_NAME
// GIT_CLONE_COMMIT_AUTHOR_EMAIL
// GIT_CLONE_COMMIT_COMMITER_NAME
// GIT_CLONE_COMMIT_COMMITER_EMAIL

// https://devcenter.bitrise.io/en/references/available-environment-variables.html
