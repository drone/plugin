// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/DevanshMathur19/plugin/plugin/internal/environ"
	"golang.org/x/exp/slog"
)

// Execer executes a bitrise plugin.
type Execer struct {
	Source     string // plugin source code directory
	Workdir    string // pipeline working directory (aka workspace)
	Environ    []string
	Stdout     io.Writer
	Stderr     io.Writer
	OutputFile string
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
		e.installAptDeps(out)
	}

	// install darwin dependencies
	if runtime.GOOS == "darwin" {
		e.installBrewDeps(out)
	}

	// create the .envstore.yml file if not present
	if !exists(e.Source, envStoreFile) {
		slog.Debug("envman init")
		cmd := exec.Command("envman", "init")
		cmd.Dir = e.Source
		if err := cmd.Run(); err != nil {
			slog.Warn("Unable to create envstore file", err)
		}
	}

	module := out.Toolkit.Go.Module
	if module == "" {
		module = out.Toolkit.Go.PackageName
	}
	// execute the plugin. the execution logic differs
	// based on programming language.
	stepEnv := e.getStepEnv(out)
	if module != "" {
		// if the plugin is a Go module
		if err := e.runGoModule(module, stepEnv); err != nil {
			return err
		}
	} else {
		// else if the plugin is a Bash script
		if err := e.runBashScript(out, stepEnv); err != nil {
			return err
		}
	}

	// save to outputfile if present
	if len(e.OutputFile) > 0 {
		if m, err := readEnvStore(e.Source); err == nil && len(m.Envs) > 0 {
			if err = saveOutputFromEnvStore(m.Envs, e.OutputFile); err != nil {
				slog.Error("Unable to save output", err)
			}
		} else if err != nil {
			slog.Error("Unable to load envstore file", err)
		}
	}

	return nil
}

func (e *Execer) runGoModule(module string, env []string) error {
	slog.Debug("go build", slog.String("module", module))
	// compile the code
	binpath := filepath.Join(e.Source, "step.exe")
	cmd := exec.Command("go", "build", "-o", binpath, module)
	cmd.Env = env
	cmd.Dir = e.Source
	cmd.Stderr = e.Stderr
	cmd.Stdout = e.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	slog.Debug("go run", slog.String("module", module))

	// execute the binary
	cmd = exec.Command(binpath)
	cmd.Env = env
	cmd.Dir = e.Workdir
	cmd.Stderr = e.Stderr
	cmd.Stdout = e.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (e *Execer) runBashScript(out *spec, env []string) error {
	// determine the default script path
	script := out.Toolkit.Bash.Entryfile
	path := filepath.Join(e.Source, script)

	slog.Debug("execute", slog.String("file", script))

	// if the bash shell does not exist fallback
	// to posix shell.
	shell, err := exec.LookPath("bash")
	if err != nil {
		shell = "/bin/sh"
	}

	// execute the binary
	cmd := exec.Command(shell, path)
	cmd.Env = env
	cmd.Dir = e.Workdir
	cmd.Stderr = e.Stderr
	cmd.Stdout = e.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (e *Execer) installBrewDeps(out *spec) {
	for _, item := range out.Deps.Brew {
		slog.Debug("brew install", slog.String("package", item.Name))

		cmd := exec.Command("brew", "install", item.Name)
		cmd.Env = e.Environ
		cmd.Dir = e.Workdir
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		cmd.Run()
	}
}

func (e *Execer) installAptDeps(out *spec) {
	if len(out.Deps.Aptget) > 0 {
		slog.Debug("apt-get update")

		cmd := exec.Command("sudo", "apt-get", "update")
		cmd.Env = e.Environ
		cmd.Dir = e.Workdir
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		cmd.Run()
	}

	for _, item := range out.Deps.Aptget {
		slog.Debug("apt-get install", slog.String("package", item.Name))

		cmd := exec.Command("sudo", "apt-get", "install", item.Name)
		cmd.Env = e.Environ
		cmd.Stderr = e.Stderr
		cmd.Stdout = e.Stdout
		cmd.Run()
	}
}

func (e *Execer) getStepEnv(out *spec) []string {
	defaults := e.getDefaults(out.Inputs)
	env := environ.Map(e.Environ)
	for k, v := range defaults {
		if _, ok := env[k]; !ok {
			env[k] = v
		}
	}
	return environ.Slice(env)
}

func (e *Execer) getDefaults(inputs []map[string]interface{}) map[string]string {
	o := make(map[string]string)

	env := environ.Map(e.Environ)
	for _, in := range inputs {
		for k, v := range in {
			if k != "opts" {
				default_ := ""
				if v == nil {
					default_ = "null"
				} else {
					default_ = fmt.Sprintf("%v", v)
				}
				if default_ != "" {
					o[k] = os.Expand(default_, func(s string) string { return env[s] })
				}
			}
		}
	}
	return o
}
