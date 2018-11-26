package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeEnvironments(t *testing.T) {
	cases := []struct {
		name     string
		env1     []string
		env2     []string
		expected []string
	}{
		{
			name:     "Empty + Empty = Empty",
			env1:     []string{},
			env2:     []string{},
			expected: []string{},
		},
		{
			name:     "right identity",
			env1:     []string{"VAR1=VALUE1", "VAR2=VALUE2"},
			env2:     []string{},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2"},
		},
		{
			name:     "left identity",
			env1:     []string{},
			env2:     []string{"VAR1=VALUE1", "VAR2=VALUE2"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2"},
		},
		{
			name:     "no overlap",
			env1:     []string{"VAR1=VALUE1", "VAR2=VALUE2"},
			env2:     []string{"VAR3=VALUE3"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2", "VAR3=VALUE3"},
		},
		{
			name:     "overlap",
			env1:     []string{"VAR1=VALUE1", "VAR2=VALUE2"},
			env2:     []string{"VAR1=VALUE3", "VAR2=VALUE4"},
			expected: []string{"VAR1=VALUE3", "VAR2=VALUE4"},
		},
		{
			name:     "PATH merge",
			env1:     []string{"PATH=c:d"},
			env2:     []string{"PATH=a:b"},
			expected: []string{"PATH=a:b:c:d"},
		},
		{
			name:     "CPATH merge",
			env1:     []string{"CPATH=c:d"},
			env2:     []string{"CPATH=a:b"},
			expected: []string{"CPATH=a:b:c:d"},
		},
		{
			name:     "LD_LIBRARY_PATH merge",
			env1:     []string{"LD_LIBRARY_PATH=c:d"},
			env2:     []string{"LD_LIBRARY_PATH=a:b"},
			expected: []string{"LD_LIBRARY_PATH=a:b:c:d"},
		},
		{
			name:     "complex example",
			env1:     []string{"VAR1=VALUE1", "PATH=/bin:/sbin"},
			env2:     []string{"PATH=~/bin:~/.local/bin", "VAR2=VALUE2"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2", "PATH=~/bin:~/.local/bin:/bin:/sbin"},
		},
		{
			name:     "discard invalid environment variables",
			env1:     []string{"VAR1", "VAR2=VALUE2", "garbagelol"},
			env2:     []string{"VAR3="},
			expected: []string{"VAR2=VALUE2", "VAR3="},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeEnvironments(tt.env1, tt.env2)
			assert.ElementsMatch(t, result, tt.expected)
		})
	}
}
