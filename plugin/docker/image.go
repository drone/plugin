// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import "strings"

// Is returns true if the string can be identified as a
// docker image. This function only recognizes well-known
// docker registry urls.
func Is(s string) bool {
	// the string is a docker image if prefixed with docker://
	if strings.HasPrefix(s, "docker://") ||
		strings.HasPrefix(s, "docker+http://") ||
		strings.HasPrefix(s, "docker+https://") {
		return true
	}

	// the string is definitely not a docker image if
	// prefixed with git://
	if strings.HasPrefix(s, "git://") ||
		strings.HasPrefix(s, "git+http://") ||
		strings.HasPrefix(s, "git+https://") {
		return false
	}

	// trim http and https prefixes
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")

	// the string is a docker image if the string is
	// a well-known docker registry url.
	switch {
	// docker.io/**
	case strings.HasPrefix(s, "docker.io/"):
		return true
	// quay.io/**
	case strings.HasPrefix(s, "quay.io/"):
		return true
	// gcr.io/**
	case strings.HasPrefix(s, "gcr.io/"):
		return true
	// *.gcr.io/**
	case strings.Contains(s, ".gcr.io/"):
		return true
	// public.ecr.aws/**
	case strings.HasPrefix(s, "public.ecr.aws/"):
		return true
	// *.azurecr.io/**
	case strings.Contains(s, "azurecr.io/"):
		return true
	// *.dkr.ecr.*.amazonaws.com/**
	case strings.Contains(s, ".dkr.ecr."):
		return true
	default:
		return false
	}
}
