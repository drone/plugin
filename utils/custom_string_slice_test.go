package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomStringSliceFlag_Set(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single value",
			input:    "test",
			expected: []string{"test"},
		},
		{
			name:     "multiple values",
			input:    "test1;test2;test3",
			expected: []string{"test1", "test2", "test3"},
		},
		{
			name:     "values with spaces",
			input:    "test 1; test 2 ; test 3",
			expected: []string{"test 1", "test 2", "test 3"},
		},
		{
			name:     "empty values between delimiters",
			input:    "test1;;test2",
			expected: []string{"test1", "", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &CustomStringSliceFlag{}
			err := flag.Set(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, flag.Value)
		})
	}
}

func TestCustomStringSliceFlag_MultipleSet(t *testing.T) {
	flag := &CustomStringSliceFlag{}

	// Test multiple Set calls
	inputs := []string{
		"test1;test2",
		"test3",
		"test4;test5",
	}

	expected := []string{
		"test1", "test2", "test3", "test4", "test5",
	}

	for _, input := range inputs {
		err := flag.Set(input)
		assert.NoError(t, err)
	}

	assert.Equal(t, expected, flag.Value)
}

func TestCustomStringSliceFlag_Integration(t *testing.T) {
	flag := &CustomStringSliceFlag{}

	// Test the complete workflow
	err := flag.Set("value1;value2")
	assert.NoError(t, err)

	err = flag.Set("value3")
	assert.NoError(t, err)

	// Test GetValue
	values := flag.GetValue()
	assert.Equal(t, []string{"value1", "value2", "value3"}, values)

	// Test String
	str := flag.String()
	assert.Equal(t, "value1;value2;value3", str)
}
