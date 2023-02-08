// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/exp/slog"
)

// Execer executes a harness plugin.
type Execer struct {
	Source  string // plugin source code directory
	Workdir string // pipeline working directory (aka workspace)
	Environ []string
	Stdout  io.Writer
	Stderr  io.Writer
}

// Exec executes a bitrise plugin.
func (e *Execer) Exec(ctx context.Context) error {
	// parse the bitrise plugin yaml
	out, err := parseFile(filepath.Join(e.Source, "plugin.yml"))
	if err != nil {
		return err
	}

	// install linux dependencies
	if runtime.GOOS == "linux" {
		if len(out.Deps.Apt) > 0 {
			slog.
				Debug("apt-get update")

			cmd := exec.Command("sudo", "apt-get", "update")
			cmd.Env = e.Environ
			cmd.Dir = e.Workdir
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}

		for _, item := range out.Deps.Apt {
			slog.
				Debug("apt-get install", slog.String("package", item))

			cmd := exec.Command("sudo", "apt-get", "install", item)
			cmd.Env = e.Environ
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}
	}

	// install darwin dependencies
	if runtime.GOOS == "darwin" {
		for _, item := range out.Deps.Brew {
			slog.
				Debug("brew install", slog.String("package", item))

			cmd := exec.Command("brew", "install", item)
			cmd.Env = e.Environ
			cmd.Dir = e.Workdir
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}
	}

	// execute the plugin. the execution logic differs
	// based on programming language.
	if module := out.Run.Go.Module; module != "" {
		// if the plugin is a Go module

		slog.
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

		slog.
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
		script := out.Run.Bash.Path
		shell := "/bin/bash"
		path := filepath.Join(e.Source, script)

		slog.
			Debug("execute", slog.String("file", script))

		// if the bash shell does not exist fallback
		// to posix shell.
		switch runtime.GOOS {
		case "windows":
			// TODO we may want to disable profile and interactive mode
			// when executing powershell scripts -noprofile -noninteractive
			shell = "powershell"
		case "linux", "darwin":
			// fallback to the posix shell if bash
			// is not available on the host.
			if _, err := exec.LookPath("bash"); err != nil {
				shell = "/bin/sh"
			}
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

	return nil
}
