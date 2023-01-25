// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

// Execer executes a harness plugin.
type Execer struct {
	Ref     string
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

		binpath, err := downloadFile(parsedURL)
		if err != nil {
			return err
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

		slog.Debug("go build", slog.String("module", module))

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

		slog.Debug("go run", slog.String("module", module))
	} else {
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

func downloadFile(url string) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temporary file")
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to download url: %s", url))
	}
	defer resp.Body.Close()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return "", errors.Wrap(err, "failed to write download binary to file")
	}

	return f.Name(), nil
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
	slog.FromContext(ctx).Debug(s)
}
