//go:build !windows
// +build !windows

package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeEnvironmentsPosix(t *testing.T) {
	cases := []struct {
		name     string
		env1     []string
		env2     []string
		env3     []string
		expected []string
	}{
		{
			name:     "mixed case",
			env1:     []string{"VAR1=VALUE1"},
			env2:     []string{"VAR2=VALUE2"},
			env3:     []string{"Var1=VALUE3", "Var2=VALUE4"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2", "Var1=VALUE3", "Var2=VALUE4"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeEnvironments(tt.env1, tt.env2, tt.env3)
			assert.ElementsMatch(t, result, tt.expected)
		})
	}
}
