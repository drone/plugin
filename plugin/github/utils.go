package github

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func getOutputVars(codedir, name string) []string {
	outputVars := make([]string, 0)

	_, _, relPath, _, err := parseActionName(name)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse action name: %s with err: %v", name, err))
		return outputVars
	}
	actionYmlFilePath := filepath.Join(codedir, relPath)
	spec, err := parseFile(getActionYamlFname(actionYmlFilePath))
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse output vars: %v", err))
	}

	if spec != nil && spec.Outputs != nil {
		for k := range spec.Outputs {
			outputVars = append(outputVars, k)
		}
	}
	return outputVars
}

func getSecretFile(envVars map[string]string, tmpDir string) (string, error) {
	v, ok := envVars["GITHUB_TOKEN"]
	if !ok {
		return "", nil
	}

	secrets := make(map[string]string)
	secrets["GITHUB_TOKEN"] = v

	secretFile := filepath.Join(tmpDir, "wf.secrets")
	if err := godotenv.Write(secrets, secretFile); err != nil {
		return "", err
	}
	return secretFile, nil
}
