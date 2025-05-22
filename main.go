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
	"github.com/drone/plugin/utils"
)

var (
	name          string                      // plugin name
	repo          string                      // plugin repository
	ref           string                      // plugin repository reference
	args          string                      // plugin arguments
	sha           string                      // plugin repository commit
	kind          string                      // plugin kind (action, bitrise, harness)
	downloadOnly  bool                        // plugin won't be executed on setting this flag. Only source will be downloaded. Used for caching the plugin dependencies
	disableClone  bool                        // plugin does not clone when this flag is enabled
	binarySources utils.CustomStringSliceFlag // plugin uses these binary source urls in the same order to download the binaires
)

func main() {
	ctx := context.Background()

	level := slog.LevelInfo
	if os.Getenv("DRONE_DEBUG") == "true" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// parse the input parameters
	flag.StringVar(&name, "name", "", "plugin name")
	flag.StringVar(&repo, "repo", "", "plugin repository")
	flag.StringVar(&ref, "ref", "", "plugin reference")
	flag.StringVar(&args, "args", "", "plugin arguments")
	flag.StringVar(&sha, "sha", "", "plugin commit")
	flag.StringVar(&kind, "kind", "", "plugin kind")
	flag.BoolVar(&downloadOnly, "download-only", false, "plugin downloadOnly")
	flag.BoolVar(&disableClone, "disable-clone", false, "disable clone functionality")
	flag.Var(&binarySources, "sources", "source urls to download binaries")
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
		repo_, ref_, sha_, ok := harness.ParseLookup(name)
		if ok {
			repo = repo_
			ref = ref_
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
		slog.Error("cannot get workdir", "error", err)
		os.Exit(1)
	}

	// clone the plugin repository
	var codedir string
	if !disableClone {
		clone := cloner.NewCache(cloner.NewDefault())
		codedir, err = clone.Clone(ctx, repo, ref, sha)
		if err != nil {
			slog.Error("cannot clone the plugin", "error", err)
			os.Exit(1)
		}
	}

	outputFile := os.Getenv("DRONE_OUTPUT")

	switch {
	// execute harness plugin
	case kind == "harness" || (kind == "" && harness.Is(codedir)):
		if !disableClone {
			slog.Info("detected harness plugin.yml")
		} else {
			slog.Info("clone is disabled for harness plugin")
		}
		execer := harness.Execer{
			Source:        codedir,
			Workdir:       workdir,
			Ref:           ref,
			Environ:       os.Environ(),
			Stdout:        os.Stdout,
			Stderr:        os.Stderr,
			BinarySources: binarySources,
			DisableClone:  disableClone,
			DownloadOnly:  downloadOnly,
			Args:          args,
		}
		if err := execer.Exec(ctx); err != nil {
			slog.Error("step failed", "error", err)
			os.Exit(1)
		}

	// execute bitrise plugin
	case kind == "bitrise" || (kind == "" && bitrise.Is(codedir)):
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
			slog.Error("step failed", "error", err)
			os.Exit(1)
		}

	case github.Is(codedir) || kind == "action":
		slog.Info("detected github action action.yml")
		execer := github.Execer{
			Name:       name,
			Source:     codedir,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Environ:    github.Environ(os.Environ()),
			OutputFile: outputFile,
		}
		if err := execer.Exec(ctx); err != nil {
			slog.Error("action step failed", "error", err)
			os.Exit(1)
		}
	default:
		slog.Info("unknown plugin type")
		os.Exit(1)
	}
}
