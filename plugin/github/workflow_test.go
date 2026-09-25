// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputVarWinScriptCreatesParentDir(t *testing.T) {
	outputFile := `C:/Windows/TEMP/engine/step-output.env`
	got := outputVarWinScript([]string{"msbuildPath"}, stepId, outputFile)

	if !strings.Contains(got, "os.makedirs") {
		t.Fatalf("expected os.makedirs in windows output script, got %q", got)
	}
	if !strings.Contains(got, outputFile) {
		t.Fatalf("expected output path %q in script, got %q", outputFile, got)
	}
}

func TestEnsureOutputFileDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "engine")
	outputFile := filepath.Join(dir, "step-output.env")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected missing parent dir before ensure, err=%v", err)
	}
	if err := ensureOutputFileDir(outputFile); err != nil {
		t.Fatalf("ensureOutputFileDir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected parent dir to exist, err=%v", err)
	}
}

func TestEnsureOutputFileDirEmpty(t *testing.T) {
	if err := ensureOutputFileDir(""); err != nil {
		t.Fatalf("empty output file should be a no-op: %v", err)
	}
}
