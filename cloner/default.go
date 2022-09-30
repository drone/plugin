// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloner

import (
	"context"
	"io"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Default returns the default cloner.
func Default() Cloner {
	return &cloner{
		depth:  1,
		remote: "origin",
		stdout: os.Stdout,
	}
}

// default cloner using the built-in Git client.
type cloner struct {
	depth  int
	remote string
	stdout io.Writer
}

// Clone the repository using the built-in Git client.
func (c *cloner) Clone(ctx context.Context, params Params) error {
	opts := &git.CloneOptions{
		Depth:      c.depth,
		Progress:   c.stdout,
		RemoteName: c.remote,
		URL:        params.Repo,
	}
	// set the reference name if provided
	if params.Ref != "" {
		opts.ReferenceName = plumbing.ReferenceName(expandRef(params.Ref))
	}
	// disable depth if cloning a specific commit sha
	if params.Sha == "" {
		opts.Depth = c.depth
	}

	// clone the repository
	r, err := git.PlainClone(params.Dir, false, opts)
	if err != nil {
		return err
	}
	if params.Sha == "" {
		return nil
	}

	// checkout the sha
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(params.Sha),
	})
}
