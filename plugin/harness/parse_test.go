// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		validate func(*testing.T, *spec)
	}{
		{
			name:     "valid plugin with binary source",
			filename: "binary_source.yml",
			wantErr:  false,
			validate: func(t *testing.T, s *spec) {
				wantSource := "https://github.com/drone-plugins/plugin/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst"
				wantFallback := ""

				if got := s.Run.Binary.Source; got != wantSource {
					t.Errorf("Expected source %q, got %q", wantSource, got)
				}
				if got := s.Run.Binary.FallbackSource; got != wantFallback {
					t.Errorf("Expected fallback source %q, got %q", wantFallback, got)
				}
			},
		},
		{
			name:     "valid plugin with binary source and fallback",
			filename: "binary_source_fallback.yml",
			wantErr:  false,
			validate: func(t *testing.T, s *spec) {
				wantSource := "https://github.com/drone-plugins/plugin/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst"
				wantFallback := "https://mirror.example.com/plugin/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst"

				if got := s.Run.Binary.Source; got != wantSource {
					t.Errorf("Expected source %q, got %q", wantSource, got)
				}
				if got := s.Run.Binary.FallbackSource; got != wantFallback {
					t.Errorf("Expected fallback source %q, got %q", wantFallback, got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use testdata directory
			testFile := filepath.Join("testdata", tt.filename)
			got, err := parseFile(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.validate(t, got)
			}
		})
	}
}

func TestParseBinaryWithFallback(t *testing.T) {
	yaml := `
run:
  binary:
    source: https://github.com/drone-plugins/drone-meltwater-cache/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst
    fallback_source: https://backup-mirror.example.com/drone-plugins/drone-meltwater-cache/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst
`
	out, err := parseString(yaml)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := out.Run.Binary.Source, "https://github.com/drone-plugins/drone-meltwater-cache/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst"; got != want {
		t.Errorf("Want source URL %q, got %q", want, got)
	}
	if got, want := out.Run.Binary.FallbackSource, "https://backup-mirror.example.com/drone-plugins/drone-meltwater-cache/releases/download/{{ release }}/plugin-{{ os }}-{{ arch }}.zst"; got != want {
		t.Errorf("Want fallback source URL %q, got %q", want, got)
	}
}
