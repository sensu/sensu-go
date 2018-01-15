package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFieldName(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"id", "ID"},
		{"myId", "MyID"},
		{"myID", "MyID"},
		{"myIdTest", "MyIDTest"},
		{"myIDTest", "MyIDTest"},
		{"idTest", "IDTest"},
		{"IdTest", "IDTest"},
		{"IDTest", "IDTest"},
		{"myIdentifier", "MyIdentifier"},
		{"myIdentifierTest", "MyIdentifierTest"},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			out := toFieldName(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
