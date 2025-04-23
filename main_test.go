package main

import (
	"flag"
	"os"
	"testing"

	"github.com/drone/plugin/utils"
	"github.com/stretchr/testify/assert"
)

// Save original command line arguments and flags
func saveOriginalFlags() []string {
	oldArgs := make([]string, len(os.Args))
	copy(oldArgs, os.Args)
	return oldArgs
}

// Restore original command line arguments and flags
func restoreOriginalFlags(oldArgs []string) {
	os.Args = make([]string, len(oldArgs))
	copy(os.Args, oldArgs)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestFlagParsing(t *testing.T) {
	// Save original flags
	oldArgs := saveOriginalFlags()
	defer restoreOriginalFlags(oldArgs)

	tests := []struct {
		name          string
		args          []string
		expectedFlags struct {
			name          string
			repo          string
			ref           string
			sha           string
			kind          string
			downloadOnly  bool
			disableClone  bool
			binarySources []string
		}
	}{
		{
			name: "all flags set",
			args: []string{
				"cmd",
				"-name", "test-plugin",
				"-repo", "github.com/test/repo",
				"-ref", "main",
				"-sha", "abc123",
				"-kind", "harness",
				"-download-only",
				"-disable-clone",
				"-sources", "source1;source2;source3",
			},
			expectedFlags: struct {
				name          string
				repo          string
				ref           string
				sha           string
				kind          string
				downloadOnly  bool
				disableClone  bool
				binarySources []string
			}{
				name:          "test-plugin",
				repo:          "github.com/test/repo",
				ref:           "main",
				sha:           "abc123",
				kind:          "harness",
				downloadOnly:  true,
				disableClone:  true,
				binarySources: []string{"source1", "source2", "source3"},
			},
		},
		{
			name: "minimal flags",
			args: []string{
				"cmd",
				"-name", "minimal-plugin",
				"-repo", "github.com/test/minimal",
			},
			expectedFlags: struct {
				name          string
				repo          string
				ref           string
				sha           string
				kind          string
				downloadOnly  bool
				disableClone  bool
				binarySources []string
			}{
				name:          "minimal-plugin",
				repo:          "github.com/test/minimal",
				ref:           "",
				sha:           "",
				kind:          "",
				downloadOnly:  false,
				disableClone:  false,
				binarySources: []string{},
			},
		},
		{
			name: "multiple sources",
			args: []string{
				"cmd",
				"-name", "source-plugin",
				"-sources", "source1;source2",
				"-sources", "source3;source4",
			},
			expectedFlags: struct {
				name          string
				repo          string
				ref           string
				sha           string
				kind          string
				downloadOnly  bool
				disableClone  bool
				binarySources []string
			}{
				name:          "source-plugin",
				repo:          "",
				ref:           "",
				sha:           "",
				kind:          "",
				downloadOnly:  false,
				disableClone:  false,
				binarySources: []string{"source1", "source2", "source3", "source4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args

			// Reset global variables
			name = ""
			repo = ""
			ref = ""
			sha = ""
			kind = ""
			downloadOnly = false
			disableClone = false
			binarySources = utils.CustomStringSliceFlag{}

			// Parse flags
			flag.StringVar(&name, "name", "", "plugin name")
			flag.StringVar(&repo, "repo", "", "plugin repository")
			flag.StringVar(&ref, "ref", "", "plugin reference")
			flag.StringVar(&sha, "sha", "", "plugin commit")
			flag.StringVar(&kind, "kind", "", "plugin kind")
			flag.BoolVar(&downloadOnly, "download-only", false, "plugin downloadOnly")
			flag.BoolVar(&disableClone, "disable-clone", false, "disable clone functionality")
			flag.Var(&binarySources, "sources", "source urls to download binaries")
			flag.Parse()

			// Assert flag values
			assert.Equal(t, tt.expectedFlags.name, name)
			assert.Equal(t, tt.expectedFlags.repo, repo)
			assert.Equal(t, tt.expectedFlags.ref, ref)
			assert.Equal(t, tt.expectedFlags.sha, sha)
			assert.Equal(t, tt.expectedFlags.kind, kind)
			assert.Equal(t, tt.expectedFlags.downloadOnly, downloadOnly)
			assert.Equal(t, tt.expectedFlags.disableClone, disableClone)
			assert.Equal(t, tt.expectedFlags.binarySources, binarySources.GetValue())
		})
	}
}
