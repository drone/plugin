// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitrise

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func TestSaveOutputFromEnvStoreCreatesParentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "engine")
	outputFile := filepath.Join(dir, "step-output.env")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected missing parent dir before write, err=%v", err)
	}

	envs := []map[string]string{{"FOO": "bar"}}
	if err := saveOutputFromEnvStore(envs, outputFile); err != nil {
		t.Fatalf("saveOutputFromEnvStore: %v", err)
	}

	got, err := godotenv.Read(outputFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if got["FOO"] != "bar" {
		t.Fatalf("expected FOO=bar, got %#v", got)
	}
}

func TestEnsureOutputFileDirEmpty(t *testing.T) {
	if err := ensureOutputFileDir(""); err != nil {
		t.Fatalf("empty output file should be a no-op: %v", err)
	}
}
