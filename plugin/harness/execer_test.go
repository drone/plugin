package harness

import (
	"testing"

	"github.com/drone/plugin/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetBinarySources(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		fallback      string
		binarySources []string
		expected      []string
	}{
		{
			name:          "all sources present",
			source:        "main-source",
			fallback:      "fallback-source",
			binarySources: []string{"binary1", "binary2"},
			expected:      []string{"binary1", "binary2", "main-source", "fallback-source"},
		},
		{
			name:          "only main source",
			source:        "main-source",
			fallback:      "",
			binarySources: []string{},
			expected:      []string{"main-source"},
		},
		{
			name:          "only binary sources",
			source:        "",
			fallback:      "",
			binarySources: []string{"binary1", "binary2"},
			expected:      []string{"binary1", "binary2"},
		},
		{
			name:          "empty sources",
			source:        "",
			fallback:      "",
			binarySources: []string{},
			expected:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execer := &Execer{
				BinarySources: utils.CustomStringSliceFlag{Value: tt.binarySources},
			}
			result := execer.getBinarySources(tt.source, tt.fallback)
			assert.Equal(t, tt.expected, result)
		})
	}
}
