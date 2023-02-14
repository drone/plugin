// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/exp/slog"
)

// Execer executes a bitrise plugin.
type Execer struct {
	Source     string // plugin source code directory
	Workdir    string // pipeline working directory (aka workspace)
	Environ    []string
	Stdout     io.Writer
	Stderr     io.Writer
	Outputfile string
}

// Exec executes a bitrise plugin.
func (e *Execer) Exec(ctx context.Context) error {
	// parse the bitrise plugin yaml
	out, err := parseFile(filepath.Join(e.Source, "step.yml"))
	if err != nil {
		return err
	}

	// install linux dependencies
	if runtime.GOOS == "linux" {
		if len(out.Deps.Aptget) > 0 {
			slog.FromContext(ctx).
				Debug("apt-get update")

			cmd := exec.Command("sudo", "apt-get", "update")
			cmd.Env = e.Environ
			cmd.Dir = e.Workdir
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}

		for _, item := range out.Deps.Aptget {
			slog.FromContext(ctx).
				Debug("apt-get install", slog.String("package", item.Name))

			cmd := exec.Command("sudo", "apt-get", "install", item.Name)
			cmd.Env = e.Environ
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}
	}

	// install darwin dependencies
	if runtime.GOOS == "darwin" {
		for _, item := range out.Deps.Brew {
			slog.FromContext(ctx).
				Debug("brew install", slog.String("package", item.Name))

			cmd := exec.Command("brew", "install", item.Name)
			cmd.Env = e.Environ
			cmd.Dir = e.Workdir
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}
	}

	// create the .envstore.yml file if not present
	if !Is(e.Source, envStoreFile) {
		slog.FromContext(ctx).
			Debug("envman init")
		cmd := exec.Command("envman", "init")
		cmd.Dir = e.Source
		if err := cmd.Run(); err != nil {
			slog.FromContext(ctx).Warn("Unable to create envstore file", err)
		}
	}

	module := out.Toolkit.Go.Module
	if module == "" {
		module = out.Toolkit.Go.PackageName
	}
	// execute the plugin. the execution logic differs
	// based on programming language.
	if module != "" {
		// if the plugin is a Go module

		slog.FromContext(ctx).
			Debug("go build", slog.String("module", module))

		// compile the code
		binpath := filepath.Join(e.Source, "step.exe")
		cmd := exec.Command("go", "build", "-o", binpath, module)
		cmd.Env = e.Environ
		cmd.Dir = e.Source
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}

		slog.FromContext(ctx).
			Debug("go run", slog.String("module", module))

		// execute the binary
		cmd = exec.Command(binpath)
		cmd.Env = e.Environ
		cmd.Dir = e.Workdir
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}

	} else {
		// else if the plugin is a Bash script

		// determine the default script path
		script := out.Toolkit.Bash.Entryfile
		path := filepath.Join(e.Source, script)

		slog.FromContext(ctx).
			Debug("execute", slog.String("file", script))

		// if the bash shell does not exist fallback
		// to posix shell.
		shell, err := exec.LookPath("bash")
		if err != nil {
			shell = "/bin/sh"
		}

		// execute the binary
		cmd := exec.Command(shell, path)
		cmd.Env = e.Environ
		cmd.Dir = e.Workdir
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// save to outputfile if present
	if len(e.Outputfile) > 0 {
		if m, err := readEnvStore(e.Source); err == nil && len(m.Envs) > 0 {
			if err = saveOutputFromEnvStore(m.Envs, e.Outputfile); err != nil {
				slog.FromContext(ctx).Error("Unable to save output", err)
			}
		} else if err != nil {
			slog.FromContext(ctx).Error("Unable to load envstore file", err)
		}
	}

	return nil
}
