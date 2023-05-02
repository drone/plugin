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
	"golang.org/x/exp/slog"
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
	log := slog.Default()

	script, err := dotenvScript(envFile)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to create pre/post-step script: %s", err))
		script = "--version"
	}

	var cmd string
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		cmd = fmt.Sprintf("python3 %s", script)
	} else {
		cmd = fmt.Sprintf("python %s", script)
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
		cmd += fmt.Sprintf("print('%s'+'='+'${{ steps.%s.outputs.%s }}'); ", outputVar, prevStepId, outputVar)
	}

	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		cmd = fmt.Sprintf("python3 -c \"%s\" > %s", cmd, outputFile)
	} else {
		cmd = fmt.Sprintf("python -c \"%s\" > %s", cmd, outputFile)
	}
	fmt.Println("output var command")
	fmt.Println(cmd)
	fmt.Println("====")
	fmt.Println(outputFile)
	fmt.Println("====")

	script := outputVarScript(outputVars, prevStepId, outputFile)
	fmt.Println("output var script")
	fmt.Println(script)
	fmt.Println("====")

	s := step{
		Name:  "output variables",
		Run:   script,
		If:    fmt.Sprintf("%t", !skip),
		Shell: "python",
	}
	// if runtime.GOOS == "windows" {
	// 	s.Shell = "powershell"
	// }
	return s
}

func dotenvScript(envFile string) (string, error) {
	script := fmt.Sprintf(`
import os
import base64

out = ""
for k, v in os.environ.items():
	if "(" not in k and ")" not in k:
		out = out + "{}={}\n".format(k, str(base64.urlsafe_b64encode(bytes(v, "utf-8")), "utf-8"))
with open(r"%s", "wb") as text_file:
	text_file.write(bytes(out, "UTF-8"))
`, envFile)

	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	file.WriteString(script)
	return file.Name(), nil
}

func outputVarScript(outputVars []string, prevStepId, outputFile string) string {
	cmd := "out = \"\"\n"
	for _, outputVar := range outputVars {
		cmd += fmt.Sprintf("out = out + \"%s=${{ steps.%s.outputs.%s }}\\n\"\n", outputVar, stepId, outputVar)
	}

	fmt.Println(outputFile)
	return fmt.Sprintf(`
%s
with open(r"%s", "wb") as text_file:
	text_file.write(bytes(out, "UTF-8"))
`, cmd, outputFile)
}
