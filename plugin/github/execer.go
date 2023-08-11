// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/drone/plugin/plugin/internal/environ"
	"github.com/nektos/act/cmd"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

// Execer executes a github action.
type Execer struct {
	Name       string
	Environ    []string
	Source     string
	OutputFile string
	Stdout     io.Writer
	Stderr     io.Writer
}

// Exec executes a github action.
func (e *Execer) Exec(ctx context.Context) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	envVars := environ.Map(e.Environ)
	envVars["RUNNER_TEMP"] = tmpDir
	outputVars := getOutputVars(e.Source, e.Name)

	workflowFile := filepath.Join(tmpDir, "workflow.yml")
	beforeStepEnvFile := filepath.Join(tmpDir, "before.env")
	afterStepEnvFile := filepath.Join(tmpDir, "after.env")
	if err := createWorkflowFile(e.Name, envVars, workflowFile, beforeStepEnvFile, afterStepEnvFile, e.OutputFile, outputVars); err != nil {
		return err
	}

	oldOsArgs := os.Args
	defer func() { os.Args = oldOsArgs }()

	os.Args = []string{
		"action",
		"-W",
		workflowFile,
		"-P",
		"-self-hosted=-self-hosted",
		"-b",
		"--detect-event",
	}

	if sFile, err := getSecretFile(envVars, tmpDir); err == nil && sFile != "" {
		os.Args = append(os.Args, "--secret-file", sFile)
	}

	if eventPayload, ok := envVars["PLUGIN_EVENT_PAYLOAD"]; ok {
		if isJSON(eventPayload) {
			eventPayloadFile := filepath.Join(tmpDir, "event.yml")

			if err := ioutil.WriteFile(eventPayloadFile, []byte(eventPayload), 0644); err != nil {
				return errors.Wrap(err, "failed to write event payload to file")
			}

			os.Args = append(os.Args, "--eventpath", eventPayloadFile)
		} else {
			slog.Debug("invalid event payload", eventPayload)
		}
	}

	cmd.Execute(ctx, "1.1")

	if err := exportEnv(beforeStepEnvFile, afterStepEnvFile); err != nil {
		return err
	}
	return nil
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
