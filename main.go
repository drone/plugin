// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"os"

	"golang.org/x/exp/slog"

	"github.com/drone/plugin/cloner"
	"github.com/drone/plugin/plugin/bitrise"
	"github.com/drone/plugin/plugin/github"
	"github.com/drone/plugin/plugin/harness"
)

var (
	name       string // plugin name
	repo       string // plugin repository
	ref        string // plugin repository reference
	sha        string // plugin repository commit
	kind       string // plugin kind (action, bitrise, harness)
	outputFile string // plugin outputfile
)

func main() {
	ctx := context.Background()

	// parse the input parameters
	flag.StringVar(&name, "name", "", "plugin name")
	flag.StringVar(&repo, "repo", "", "plugin repository")
	flag.StringVar(&ref, "ref", "", "plugin reference")
	flag.StringVar(&sha, "sha", "", "plugin commit")
	flag.StringVar(&kind, "kind", "", "plugin kind")
	flag.Parse()

	// the user may specific the action plugin alias instead
	// of the git repository. We are able to lookup the plugin
	// by alias to find the corresponding repository and ref.
	if repo == "" && kind == "action" {
		repo_, ref_, ok := github.ParseLookup(name)
		if ok {
			repo = repo_
			ref = ref_
		}
	}

	// the user may specific the harness plugin alias instead
	// of the git repository. We are able to lookup the plugin
	// by alias to find the corresponding repository and commit.
	if repo == "" && kind == "harness" {
		repo_, sha_, ok := harness.ParseLookup(name)
		if ok {
			repo = repo_
			sha = sha_
		}
	}

	// the user may specific the bitrise plugin alias instead
	// of the git repository. We are able to lookup the plugin
	// by alias to find the corresponding repository and commit.
	if repo == "" && kind == "bitrise" {
		repo_, sha_, ok := bitrise.ParseLookup(name)
		if ok {
			repo = repo_
			sha = sha_
		}
	}

	// current working directory (workspace)
	workdir, err := os.Getwd()
	if err != nil {
		slog.Error("cannot get workdir", err)
		os.Exit(1)
	}

	// directory to clone the plugin
	codedir, err := os.MkdirTemp("", "")
	if err != nil {
		slog.Error("cannot create clone dir", err)
		os.Exit(1)
	}
	// remove the temporary clone directory
	// after execution.
	defer os.RemoveAll(codedir)

	// clone the plugin repository
	clone := cloner.NewDefault()
	err = clone.Clone(ctx, cloner.Params{
		Repo: repo,
		Ref:  ref,
		Sha:  sha,
		Dir:  codedir,
	})
	if err != nil {
		slog.Error("cannot clone the plugin", err)
		os.Exit(1)
	}

	outputFile = os.Getenv("DRONE_OUTPUT")

	switch {
	// execute harness plugin
	case harness.Is(codedir) || kind == "harness":
		slog.Info("detected harness plugin.yml")
		execer := harness.Execer{
			Source:  codedir,
			Workdir: workdir,
			Ref:     ref,
			Environ: os.Environ(),
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		}
		if err := execer.Exec(ctx); err != nil {
			slog.Error("step failed", err)
			os.Exit(1)
		}

	// execute bitrise plugin
	case bitrise.Is(codedir) || kind == "bitrise":
		slog.Info("detected bitrise step.yml")
		execer := bitrise.Execer{
			Source:  codedir,
			Workdir: workdir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Environ: bitrise.Environ(
				os.Environ(),
			),
			OutputFile: outputFile,
		}
		if err := execer.Exec(ctx); err != nil {
			slog.Error("step failed", err)
			os.Exit(1)
		}

	case github.Is(codedir) || kind == "action":
		slog.Info("detected github action action.yml")
		execer := github.Execer{
			Name:       name,
			Source:     codedir,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Environ:    os.Environ(),
			OutputFile: outputFile,
		}
		if err := execer.Exec(ctx); err != nil {
			slog.Error("action step failed", err)
			os.Exit(1)
		}
	default:
		slog.Info("unknown plugin type")
		os.Exit(1)
	}
}
