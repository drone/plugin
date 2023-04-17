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

	// install dependencies
	if runtime.GOOS == "linux" {
		e.installAptDeps(ctx, out.Deps.Apt.Packages, out.Deps.Apt.Sources)
	} else if runtime.GOOS == "darwin" {
		e.installBrewDeps(ctx, out.Deps.Brew)
	} else if runtime.GOOS == "windows" {
		e.installChocoDeps(ctx, out.Deps.Choco)
	}

	// execute the plugin. the execution logic differs
	// based on programming language.
	if source := out.Run.Binary.Source; source != "" {
		return e.runSourceExecutable(ctx, out.Run.Binary.Source)
	} else if module := out.Run.Go.Module; module != "" {
		return e.runGoExecutable(ctx, module)
	} else {
		return e.runShellExecutable(ctx, out)
	}
}

func (e *Execer) runSourceExecutable(ctx context.Context, source string) error {
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
	return runCmds(ctx, cmds, e.Environ, e.Workdir, e.Stdout, e.Stderr)
}

func (e *Execer) runGoExecutable(ctx context.Context, module string) error {
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
	return runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir, e.Stdout, e.Stderr)
}

func (e *Execer) runShellExecutable(ctx context.Context, out *spec) error {
	if e.DownloadOnly {
		slog.Info("Download only flag is set. Not executing the plugin")
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		// TODO we may want to disable profile and interactive mode
		// when executing powershell scripts -noprofile -noninteractive
		path := filepath.Join(e.Source, out.Run.Pwsh.Path)
		slog.Debug("execute", slog.String("file", path))
		script := fmt.Sprintf(
			"$ErrorActionPreference = 'Stop'; $ProgressPreference = 'SilentlyContinue'; %s", path)
		cmd := exec.Command("pwsh", "-Command", script)
		return runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir, e.Stdout, e.Stderr)
	case "linux", "darwin":
		path := filepath.Join(e.Source, out.Run.Bash.Path)

		// fallback to the posix shell if bash
		// is not available on the host.
		shell := "bash"
		if _, err := exec.LookPath("bash"); err != nil {
			shell = "sh"
		}
		slog.Debug("execute", slog.String("file", path))

		cmd := exec.Command(shell, path)
		return runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir, e.Stdout, e.Stderr)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
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

func (e *Execer) installAptDeps(ctx context.Context, deps []string, sources []*AptSource) {
	for _, source := range sources {
		if source == nil || source.Key == "" || source.Data == "" {
			slog.Info("Either key or data is not set", slog.String("source", source.Data), slog.String("key", source.Key))
			continue
		}

		cmdStr := fmt.Sprintf("wget -qO - %s | sudo apt-key add - echo \"%s\" | sudo tee -a /etc/apt/sources.list", source.Key, source.Data)
		fmt.Println(cmdStr)
		cmd := exec.Command("bash", "-c", cmdStr)
		if err := runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir,
			e.Stdout, e.Stderr); err != nil {
			slog.Error("add-apt-repository failed", "error", err)
		}
	}

	if len(deps) > 0 {
		slog.Debug("apt-get update")

		cmd := exec.Command("sudo", "apt-get", "update")
		if err := runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir,
			e.Stdout, e.Stderr); err != nil {
			slog.Error("apt-get update failed", "error", err)
		}

	}

	for _, item := range deps {
		slog.Debug("apt-get install", slog.String("package", item))

		cmd := exec.Command("sudo", "apt-get", "install", item)
		if err := runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir,
			e.Stdout, e.Stderr); err != nil {
			slog.Error("apt-get install failed", slog.String("package", item), "error", err)
		}
	}
}

func (e *Execer) installBrewDeps(ctx context.Context, deps []string) {
	for _, item := range deps {
		slog.Debug("brew install", slog.String("package", item))

		cmd := exec.Command("brew", "install", item)
		if err := runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir,
			e.Stdout, e.Stderr); err != nil {
			slog.Error("brew install failed", slog.String("package", item), "error", err)
		}
	}
}

func (e *Execer) installChocoDeps(ctx context.Context, deps []string) {
	for _, item := range deps {
		slog.Debug("choco install", slog.String("package", item))

		cmd := exec.Command("choco", "install", item)
		if err := runCmds(ctx, []*exec.Cmd{cmd}, e.Environ, e.Workdir,
			e.Stdout, e.Stderr); err != nil {
			slog.Error("choco install failed", slog.String("package", item), "error", err)
		}
	}
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
