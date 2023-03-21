// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/drone/plugin/cache"
	"github.com/drone/plugin/plugin/internal/file"
	"golang.org/x/exp/slog"
)

// Execer executes a harness plugin.
type Execer struct {
	Ref          string // Git ref for source code
	Source       string // plugin source code directory
	Workdir      string // pipeline working directory (aka workspace)
	DownloadOnly bool
	Environ      []string
	Stdout       io.Writer
	Stderr       io.Writer
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
			slog.Debug("apt-get update")

			cmd := exec.Command("sudo", "apt-get", "update")
			cmd.Env = e.Environ
			cmd.Dir = e.Workdir
			cmd.Stderr = e.Stderr
			cmd.Stdout = e.Stdout
			cmd.Run()
		}

		for _, item := range out.Deps.Apt {
			slog.Debug("apt-get install", slog.String("package", item))

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
			slog.Debug("brew install", slog.String("package", item))

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
	if source := out.Run.Binary.Source; source != "" {
		parsedURL, err := NewMetadata(source, e.Ref).Generate()
		if err != nil {
			return err
		}
		binpath, err := file.Download(parsedURL)
		if err != nil {
			return err
		}

		if e.DownloadOnly {
			slog.Info("Download only flag is set. Not executing the plugin")
			return nil
		}

		var cmds []*exec.Cmd
		if runtime.GOOS != "windows" {
			cmds = append(cmds, exec.Command("chmod", "+x", binpath))
		}
		cmds = append(cmds, exec.Command(binpath))
		err = runCmds(ctx, cmds, e.Environ, e.Workdir, e.Stdout, e.Stderr)
		if err != nil {
			return err
		}
	} else if module := out.Run.Go.Module; module != "" {
		// if the plugin is a Go module
		binpath, err := e.buildGoExecutable(ctx, module)
		if err != nil {
			return err
		}

		if e.DownloadOnly {
			slog.Info("Download only flag is set. Not executing the plugin")
			return nil
		}

		slog.Debug("go run", slog.String("module", module))
		// execute the binary
		cmd := exec.Command(binpath)
		err = runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir, e.Stdout, e.Stderr)
		if err != nil {
			return err
		}
	} else {
		if e.DownloadOnly {
			slog.Info("Download only flag is set. Not executing the plugin")
			return nil
		}

		// else if the plugin is a Bash script

		// determine the default script path
		script := out.Run.Bash.Path
		shell := "/bin/bash"
		path := filepath.Join(e.Source, script)

		slog.Debug("execute", slog.String("file", script))

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
		err = runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir, e.Stdout, e.Stderr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Execer) buildGoExecutable(ctx context.Context, module string) (
	string, error) {
	defer timer("buildGoExecutable")()
	key := e.Source
	binpath := filepath.Join(e.Source, "step.exe")

	buildFn := func() error {
		slog.Debug("go build", slog.String("module", module))

		// compile the code
		cmd := exec.Command("go", "build", "-o", binpath, module)
		return runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Source, e.Stdout, e.Stderr)
	}

	if err := cache.Add(key, buildFn); err != nil {
		return "", err
	}
	return binpath, nil
}

func runCmds(ctx context.Context, cmds []*exec.Cmd, env []string, workdir string,
	stdout io.Writer, stderr io.Writer) error {
	for _, cmd := range cmds {
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Env = env
		cmd.Dir = workdir
		trace(ctx, cmd)

		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(ctx context.Context, cmd *exec.Cmd) {
	s := fmt.Sprintf("+ %s\n", strings.Join(cmd.Args, " "))
	slog.Debug(s)
}

// timer returns a function that prints the elapsed time between
// the call to timer and the call to the returned function.
// The returned function is intended to be used in a defer statement:
//
//	defer timer("sum")()
//
// Source: https://stackoverflow.com/a/45766707
func timer(name string) func() {
	start := time.Now()
	return func() {
		slog.Debug("time taken", "name", name,
			"time_secs", time.Since(start).Seconds())
	}
}
