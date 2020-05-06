package environment

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Makes platform compliant list of values
func mkList(key string, s ...string) string {
	val := strings.Join(s, string(os.PathListSeparator))
	return fmt.Sprintf("%s=%s", key, val)
}

func TestMergeEnvironments(t *testing.T) {
	cases := []struct {
		name     string
		env1     []string
		env2     []string
		env3     []string
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
			env1:     []string{mkList("PATH", "c", "d")},
			env2:     []string{mkList("PATH", "a", "b")},
			expected: []string{mkList("PATH", "a", "b", "c", "d")},
		},
		{
			name:     "CPATH merge",
			env1:     []string{mkList("CPATH", "c", "d")},
			env2:     []string{mkList("CPATH", "a", "b")},
			expected: []string{mkList("CPATH", "a", "b", "c", "d")},
		},
		{
			name:     "LD_LIBRARY_PATH merge",
			env1:     []string{mkList("LD_LIBRARY_PATH", "c", "d")},
			env2:     []string{mkList("LD_LIBRARY_PATH", "a", "b")},
			expected: []string{mkList("LD_LIBRARY_PATH", "a", "b", "c", "d")},
		},
		{
			name:     "complex example",
			env1:     []string{"VAR1=VALUE1", mkList("PATH", "/bin", "/sbin")},
			env2:     []string{mkList("PATH", "~/bin", "~/.local/bin"), "VAR2=VALUE2"},
			expected: []string{"VAR1=VALUE1", "VAR2=VALUE2", mkList("PATH", "~/bin", "~/.local/bin", "/bin", "/sbin")},
		},
		{
			name:     "discard invalid environment variables",
			env1:     []string{"VAR1", "VAR2=VALUE2", "garbagelol"},
			env2:     []string{"VAR3="},
			expected: []string{"VAR2=VALUE2", "VAR3="},
		},
		{
			name:     "more than two sets of variables",
			env1:     []string{mkList("CPATH", "e", "f"), "VAR1=two"},
			env2:     []string{mkList("CPATH", "c", "d"), "VAR1=one"},
			env3:     []string{mkList("CPATH", "a", "b")},
			expected: []string{mkList("CPATH", "a", "b", "c", "d", "e", "f"), "VAR1=one"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeEnvironments(tt.env1, tt.env2, tt.env3)
			assert.ElementsMatch(t, result, tt.expected)
		})
	}
}

func TestKey(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "special characters are replaced",
			s:    "FOO@BAR",
			want: "FOO_BAR",
		},
		{
			name: "the key is uppercase",
			s:    "foo",
			want: "FOO",
		},
		{
			name: "underscores are preserved",
			s:    "FOO_BAR",
			want: "FOO_BAR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Key(tt.s); got != tt.want {
				t.Errorf("Key() = %v, want %v", got, tt.want)
			}
		})
	}
}
