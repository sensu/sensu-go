package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractExtendedAttributes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	var m MyType

	// Empty extended attributes
	msg := []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}]}`)
	attrs, err := extractExtendedAttributes(m, msg)
	require.NoError(err)
	assert.Nil(attrs)

	// Defined extended attributes
	msg = []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}], "extendedattr": "such extended"}`)
	attrs, err = extractExtendedAttributes(m, msg)
	require.NoError(err)
	assert.Equal([]byte(`{"extendedattr":"such extended"}`), attrs)
}

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
