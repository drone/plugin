// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bitrise provides support for executing Bitrise Steps.
package bitrise

import (
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	envStoreFile = ".envstore.yml"
)

// Is returns true if the root path is a Bitrise
// plugin repository.
func Is(root string) bool {
	path := filepath.Join(root, "step.yml")
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func exists(base, file string) bool {
	path := filepath.Join(base, file)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func readEnvStore(root string) (*envStore, error) {
	buf, err := ioutil.ReadFile(filepath.Join(root, envStoreFile))
	if err != nil {
		return nil, err
	}

	m := &envStore{}
	err = yaml.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}

	return m, err
}

func saveOutputFromEnvStore(envs []map[string]string, outputfile string) error {
	finalMap := make(map[string]string)
	for _, env := range envs {
		for k, v := range env {
			finalMap[k] = v
		}
	}
	if len(finalMap) > 0 {
		return godotenv.Write(finalMap, outputfile)
	}
	return nil
}
