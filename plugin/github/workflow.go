// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type workflow struct {
	Name string         `yaml:"name"`
	On   string         `yaml:"on"`
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	Name   string `yaml:"name"`
	RunsOn string `yaml:"runs-on"`
	Steps  []step `yaml:"steps"`
}

type step struct {
	Id    string            `yaml:"id,omitempty"`
	Uses  string            `yaml:"uses,omitempty"`
	Name  string            `yaml:"name,omitempty"`
	With  map[string]string `yaml:"with,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
	Run   string            `yaml:"run,omitempty"`
	Shell string            `yaml:"shell,omitempty"`
	If    string            `yaml:"if,omitempty"`
}

const (
	workflowEvent = "push"
	workflowName  = "drone-github-action"
	jobName       = "action"
	runsOnImage   = "-self-hosted"
	stepId        = "stepId"
)

func createWorkflowFile(action string, envVars map[string]string,
	ymlFile, beforeStepEnvFile, afterStepEnvFile string, outputFile string, outputVars []string) error {
	with, err := getWith(envVars)
	if err != nil {
		return err
	}
	env := getEnv(envVars)
	j := job{
		Name:   jobName,
		RunsOn: runsOnImage,
		Steps: []step{
			prePostStep("before", beforeStepEnvFile),
			{
				Id:   stepId,
				Uses: action,
				With: with,
				Env:  env,
			},
			getOutputVariables(stepId, outputFile, outputVars),
			prePostStep("after", afterStepEnvFile),
		},
	}
	wf := &workflow{
		Name: workflowName,
		On:   getWorkflowEvent(),
		Jobs: map[string]job{
			jobName: j,
		},
	}

	out, err := yaml.Marshal(&wf)
	if err != nil {
		return errors.Wrap(err, "failed to create action workflow yml")
	}

	if err = ioutil.WriteFile(ymlFile, out, 0644); err != nil {
		return errors.Wrap(err, "failed to write yml workflow file")
	}

	return nil
}

func getWorkflowEvent() string {
	buildEvent := os.Getenv("DRONE_BUILD_EVENT")
	if buildEvent == "push" || buildEvent == "pull_request" || buildEvent == "tag" {
		return buildEvent
	}
	return "custom"
}

func prePostStep(name, envFile string) step {
	var cmd string
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		cmd = fmt.Sprintf("python3 -c 'import os; import base64; [print(k+\"=\"+str(base64.b64encode(bytes(v, \"utf-8\")), \"utf-8\")) for k, v in os.environ.items()]' > %s", envFile)
	} else {
		cmd = fmt.Sprintf("python -c 'import os; import base64; [print(k+\"=\"+str(base64.b64encode(bytes(v, \"utf-8\")), \"utf-8\")) for k, v in os.environ.items()]' > %s", envFile)
	}
	s := step{
		Name: name,
		Run:  cmd,
	}
	if runtime.GOOS == "windows" {
		s.Shell = "powershell"
	}
	return s
}

func getOutputVariables(prevStepId, outputFile string, outputVars []string) step {
	skip := len(outputFile) == 0 || len(outputVars) == 0
	cmd := ""
	for _, outputVar := range outputVars {
		cmd += fmt.Sprintf("print(\"%s\"+\"=\"+\"${{ steps.%s.outputs.%s }}\"); ", outputVar, prevStepId, outputVar)
	}

	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		cmd = fmt.Sprintf("python3 -c '%s' > %s", cmd, outputFile)
	} else {
		cmd = fmt.Sprintf("python -c '%s' > %s", cmd, outputFile)
	}
	s := step{
		Name: "output variables",
		Run:  cmd,
		If:   fmt.Sprintf("%t", !skip),
	}
	if runtime.GOOS == "windows" {
		s.Shell = "powershell"
	}
	return s
}
