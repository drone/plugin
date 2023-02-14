// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/drone/plugin/plugin/internal/environ"
	"github.com/nektos/act/cmd"
	"github.com/pkg/errors"
)

// Execer executes a github action.
type Execer struct {
	Name       string
	Environ    []string
	TmpDir     string
	Outputfile string
	Stdout     io.Writer
	Stderr     io.Writer
}

// Exec executes a github action.
func (e *Execer) Exec(ctx context.Context) error {
	envVars := environ.Map(e.Environ)
	outputVars := make([]string, 0)
	// parse the github plugin yaml
	if out, _ := parseFile(getYamlFilename(e.TmpDir)); out != nil && len(out.Outputs) > 0 {
		for k := range out.Outputs {
			outputVars = append(outputVars, k)
		}
	}

	workflowFile := filepath.Join(e.TmpDir, "workflow.yml")
	beforeStepEnvFile := filepath.Join(e.TmpDir, "before.env")
	afterStepEnvFile := filepath.Join(e.TmpDir, "after.env")
	if err := createWorkflowFile(e.Name, envVars, workflowFile, beforeStepEnvFile, afterStepEnvFile, e.Outputfile, outputVars); err != nil {
		return err
	}

	oldOsArgs := os.Args
	defer func() { os.Args = oldOsArgs }()

	os.Args = []string{
		"action",
		"-W",
		workflowFile,
		"-P",
		fmt.Sprintf("-self-hosted=-self-hosted"),
		"-b",
		"--detect-event",
	}

	if eventPayload, ok := envVars["PLUGIN_EVENT_PAYLOAD"]; ok {
		eventPayloadFile := filepath.Join(e.TmpDir, "event.yml")

		if err := ioutil.WriteFile(eventPayloadFile, []byte(eventPayload), 0644); err != nil {
			return errors.Wrap(err, "failed to write event payload to file")
		}

		os.Args = append(os.Args, "--eventpath", eventPayloadFile)
	}

	cmd.Execute(ctx, "1.1")

	if err := exportEnv(beforeStepEnvFile, afterStepEnvFile); err != nil {
		return err
	}
	return nil
}
