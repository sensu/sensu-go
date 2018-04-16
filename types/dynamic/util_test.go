package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapOfExtendedAttributes(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty input",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "simple structure",
			input: map[string]interface{}{
				"foo": "bar",
			},
			expected: map[string]interface{}{
				"Foo": "bar",
			},
		},
		{
			name: "nested structure",
			input: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
			expected: map[string]interface{}{
				"Foo": map[string]interface{}{
					"Bar": "baz",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapOfExtendedAttributes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
