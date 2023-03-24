// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClone(t *testing.T) {
	for name, tt := range map[string]struct {
		Err              error
		URL, Ref, Commit string
	}{
		"tag": {
			Err: nil,
			URL: "https://github.com/actions/checkout",
			Ref: "v2",
		},
		"branch": {
			Err: nil,
			URL: "https://github.com/anchore/scan-action",
			Ref: "act-fails",
		},
		"tag-non-sermver": {
			Err: nil,
			URL: "https://github.com/shubham149/drone-s3",
			Ref: "setup-node-and-dependencies+1.0.9",
		},
		// "sha": {
		// 	Err: nil,
		// 	URL: "https://github.com/actions/checkout",
		// 	Ref: "5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f", // v2
		// },
		// "short-sha": {
		// 	// Err: &Error{ErrShortRef, "5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f"},
		// 	URL: "https://github.com/actions/checkout",
		// 	Ref: "5a4ac90", // v2
		// },
	} {
		t.Run(name, func(t *testing.T) {
			c := NewDefault()
			err := c.Clone(context.Background(), Params{Repo: tt.URL, Ref: tt.Ref, Dir: testDir(t)})
			if tt.Err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.Err, err)
			} else {
				assert.Empty(t, err)
			}
		})
	}
}

func testDir(t *testing.T) string {
	basedir, err := os.MkdirTemp("", "clone-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(basedir) })
	return basedir
}
