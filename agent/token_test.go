package agent

import (
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
)

func TestTokenSubstitution(t *testing.T) {
	testCases := []struct {
		name          string
		data          interface{}
		input         interface{}
		expectedError bool
	}{
		{
			name:          "empty data",
			data:          &types.Entity{},
			input:         types.FixtureCheckConfig("check"),
			expectedError: false,
		},
		{
			name:          "empty input",
			data:          types.FixtureEntity("entity"),
			input:         &types.CheckConfig{},
			expectedError: false,
		},
		{
			name:          "invalid input",
			data:          types.FixtureEntity("entity"),
			input:         make(chan int),
			expectedError: true,
		},
		{
			name:          "invalid template",
			data:          types.FixtureEntity("entity"),
			input:         &types.CheckConfig{Name: "{{nil}}"},
			expectedError: true,
		},
		{
			name:          "simple template",
			data:          types.FixtureEntity("entity"),
			input:         &types.CheckConfig{Command: "{{ .ID }}"},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tokenSubstitution(tc.data, tc.input)
			testutil.CompareError(err, tc.expectedError, t)
		})
	}
}
