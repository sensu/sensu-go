package agent

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestTokenSubstitution(t *testing.T) {
	testCases := []struct {
		name            string
		data            interface{}
		input           interface{}
		expectedCommand string
		expectedError   bool
	}{
		{
			name:            "empty data",
			data:            &types.Entity{},
			input:           *types.FixtureCheckConfig("check"),
			expectedCommand: "command",
			expectedError:   false,
		},
		{
			name:            "empty input",
			data:            types.FixtureEntity("entity"),
			input:           types.CheckConfig{},
			expectedCommand: "",
			expectedError:   false,
		},
		{
			name:            "invalid input",
			data:            types.FixtureEntity("entity"),
			input:           make(chan int),
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name:            "invalid template",
			data:            types.FixtureEntity("entity"),
			input:           types.CheckConfig{Command: "{{nil}}"},
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name:            "simple template",
			data:            types.FixtureEntity("entity"),
			input:           types.CheckConfig{Command: "{{ .Name }}"},
			expectedCommand: "entity",
			expectedError:   false,
		},
		{
			name:            "default value for existing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           types.CheckConfig{Command: `{{ .Name | default "bar" }}`},
			expectedCommand: "foo",
			expectedError:   false,
		},
		{
			name:            "default value for missing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           types.CheckConfig{Command: `{{ .Check.Foo | default "bar" }}`},
			expectedCommand: "bar",
			expectedError:   false,
		},
		{
			name:            "default int value for missing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           types.CheckConfig{Command: `{{ .Check.Foo | default 1 }}`},
			expectedCommand: "1",
			expectedError:   false,
		},
		{
			name:          "unmatched token",
			data:          map[string]interface{}{"Name": "foo"},
			input:         types.CheckConfig{Command: `{{ .System.Hostname }}`},
			expectedError: true,
		},
		{
			name:            "extra escape character",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           types.CheckConfig{Command: `{{ .ID | default \"bar\" }}`},
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name: "multiple tokens and valid json",
			data: types.FixtureEntity("entity"),
			input: types.CheckConfig{Command: `{{ .Name }}; {{ "hello" }}; {{ .EntityClass }}`,
				ProxyRequests: &types.ProxyRequests{EntityAttributes: []string{`entity.EntityClass == \"proxy\"`}},
			},
			expectedCommand: "entity; hello; host",
			expectedError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TokenSubstitution(tc.data, tc.input)
			testutil.CompareError(err, tc.expectedError, t)

			if !tc.expectedError {
				checkResult := types.CheckConfig{}
				err = json.Unmarshal(result, &checkResult)
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedCommand, checkResult.Command)
			}
		})
	}
}
