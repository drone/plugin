// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/drone/plugin/cloner"
	"github.com/drone/plugin/plugin/bitrise"
	"github.com/drone/plugin/plugin/github"
	"github.com/drone/plugin/plugin/harness"

	"golang.org/x/exp/slog"
)

var (
	// plugin repository
	repo = flag.String("repo", "https://github.com/bradrydzewski/test-step.git", "")

	// plugin repository reference
	ref = flag.String("ref", "refs/heads/main", "")
)

func main() {
	log := slog.Default()
	ctx := context.Background()
	ctx = slog.NewContext(ctx, log)

	// parse the input parameters
	flag.Parse()

	// current working directory (workspace)
	workdir, err := os.Getwd()
	if err != nil {
		log.Error("cannot get workdir", err)
		os.Exit(1)
	}

	// directory to clone the plugin
	codedir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(codedir)
	if err != nil {
		log.Error("cannot create clone dir", err)
		os.Exit(1)
	}

	// clone the plugin repository
	clone := cloner.Default()
	err = clone.Clone(ctx, cloner.Params{
		Repo: *repo,
		Ref:  *ref,
		Dir:  codedir,
	})
	if err != nil {
		log.Error("cannot clone the plugin", err)
		os.Exit(1)
	}

	switch {
	// execute bitrise plugin
	case bitrise.Is(codedir):
		execer := bitrise.Execer{
			Source:  codedir,
			Workdir: workdir,
			Environ: os.Environ(),
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		}

		if err := execer.Exec(ctx); err != nil {
			os.Exit(1)
		}
	// execute harness plugin
	case harness.Is(codedir):
		// TODO

	// execute github action
	case github.Is(codedir):
		// TODO

	default:
		log.Info("unknown plugin type")
		os.Exit(1)
	}
}
