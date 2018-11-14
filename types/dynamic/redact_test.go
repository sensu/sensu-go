package dynamic

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RedactedType struct {
	A   []string               `json:",omitempty"`
	C   chan struct{}          `json:",omitempty"`
	F   func()                 `json:",omitempty"`
	I   interface{}            `json:",omitempty"`
	Int int                    `json:",omitempty"`
	M   map[string]interface{} `json:",omitempty"`
	Ptr *RedactedType          `json:",omitempty"`
	S   string                 `json:",omitempty"`
}

func (r *RedactedType) Get(name string) (interface{}, error) {
	return GetField(r, name)
}

func TestRedact(t *testing.T) {
	var nilRedactedType *RedactedType

	testCases := []struct {
		name         string
		input        interface{}
		redactFields []string
		expected     interface{}
	}{
		{
			name: "string field",
			input: &RedactedType{
				S: "foo",
			},
			redactFields: []string{"s"},
			expected: &RedactedType{
				S: Redacted,
			},
		},
		{
			name: "integer field",
			input: &RedactedType{
				Int: 42,
			},
			redactFields: []string{"int"},
			expected: &RedactedType{
				Int: 0,
			},
		},
		{
			name: "field within interface",
			input: &RedactedType{
				I: RedactedType{S: "foo"},
			},
			redactFields: []string{"S"},
			expected: &RedactedType{
				I: RedactedType{S: Redacted},
				S: Redacted,
			},
		},
		{
			name: "interface field",
			input: &RedactedType{
				I: RedactedType{S: "foo"},
			},
			redactFields: []string{"I"},
			expected: &RedactedType{
				I: interface{}(nil),
			},
		},
		{
			name: "map field",
			input: &RedactedType{
				M: map[string]interface{}{"S": "foo"},
			},
			redactFields: []string{"M"},
			expected: &RedactedType{
				M: map[string]interface{}(nil),
			},
		},
		{
			name: "field within pointer",
			input: &RedactedType{
				Ptr: &RedactedType{S: "foo"},
				S:   "foo",
			},
			redactFields: []string{"s"},
			expected: &RedactedType{
				Ptr: &RedactedType{S: Redacted},
				S:   Redacted,
			},
		},
		{
			name: "pointer field",
			input: &RedactedType{
				Ptr: &RedactedType{S: "foo"},
			},
			redactFields: []string{"ptr"},
			expected: &RedactedType{
				Ptr: nilRedactedType,
			},
		},
		{
			name: "func attribute",
			input: &RedactedType{
				F: func() { fmt.Println("function") },
			},
			redactFields: []string{"F"},
			expected:     &RedactedType{},
		},
		{
			name: "chan attribute",
			input: &RedactedType{
				C: make(chan struct{}, 1),
			},
			redactFields: []string{"C"},
			expected:     &RedactedType{},
		},
		{
			name: "slice attribute",
			input: &RedactedType{
				A: []string{"foo", "bar"},
			},
			redactFields: []string{"A"},
			expected:     &RedactedType{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Redact(tc.input, tc.redactFields...)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
