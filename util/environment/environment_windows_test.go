// +build windows

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
		env3     []string
		expected []string
	}{
		{
			name:     "no merge",
			env1:     []string{"VAR1=one"},
			env2:     []string{"VAR2=two"},
			expected: []string{"VAR1=one", "VAR2=two"},
		},
		{
			name:     "merge PATH",
			env1:     []string{"PATH=c;d"},
			env2:     []string{"PATH=a;b"},
			expected: []string{"PATH=a;b;c;d"},
		},
		{
			name:     "multiple",
			env1:     []string{"PATH=e:f"},
			env2:     []string{"PATH=c:d"},
			env3:     []string{"PATH=a:b"},
			expected: []string{"PATH=a;b;c;d;e;f"},
		},
		{
			name:     "complex example",
			env1:     []string{"VAR1=VALUE1", "PATH=/bin;/sbin"},
			env2:     []string{"PATH=~/bin;~/.local/bin", "VAR2=VALUE2"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2", "PATH=~/bin;~/.local/bin;/bin;/sbin"},
		},
		{
			name:     "mixed case",
			env1:     []string{"VAR1=VALUE1", "PATH=/bin;/sbin"},
			env2:     []string{"Path=~/bin;~/.local/bin", "VAR2=VALUE2"},
			env3:     []string{"Var1=VALUE3", "Var2=VALUE4"},
			expected: []string{"VAR1=VALUE3", "VAR2=VALUE4", "PATH=~/bin;~/.local/bin;/bin;/sbin"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeEnvironments(tt.env1, tt.env2, tt.env3)
			assert.ElementsMatch(t, result, tt.expected)
		})
	}
}
