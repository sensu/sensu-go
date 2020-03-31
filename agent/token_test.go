package agent

import (
	"encoding/json"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types/dynamic"
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
			data:            &corev2.Entity{},
			input:           *corev2.FixtureCheckConfig("check"),
			expectedCommand: "command",
			expectedError:   false,
		},
		{
			name:            "empty input",
			data:            corev2.FixtureEntity("entity"),
			input:           corev2.CheckConfig{},
			expectedCommand: "",
			expectedError:   false,
		},
		{
			name:            "invalid input",
			data:            corev2.FixtureEntity("entity"),
			input:           make(chan int),
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name:            "invalid template",
			data:            corev2.FixtureEntity("entity"),
			input:           corev2.CheckConfig{Command: "{{nil}}"},
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name:            "simple template",
			data:            corev2.FixtureEntity("entity"),
			input:           corev2.CheckConfig{Command: "{{ .name }}"},
			expectedCommand: "entity",
			expectedError:   false,
		},
		{
			name:            "default value for existing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           corev2.CheckConfig{Command: `{{ .Name | default "bar" }}`},
			expectedCommand: "foo",
			expectedError:   false,
		},
		{
			name:            "default value for missing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           corev2.CheckConfig{Command: `{{ .Check.Foo | default "bar" }}`},
			expectedCommand: "bar",
			expectedError:   false,
		},
		{
			name:            "default int value for missing field",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           corev2.CheckConfig{Command: `{{ .Check.Foo | default 1 }}`},
			expectedCommand: "1",
			expectedError:   false,
		},
		{
			name:          "unmatched token",
			data:          map[string]interface{}{"Name": "foo"},
			input:         corev2.CheckConfig{Command: `{{ .System.Hostname }}`},
			expectedError: true,
		},
		{
			name:            "extra escape character",
			data:            map[string]interface{}{"Name": "foo", "Check": map[string]interface{}{"Name": "check_foo"}},
			input:           corev2.CheckConfig{Command: `{{ .Name | default \"bar\" }}`},
			expectedCommand: "",
			expectedError:   true,
		},
		{
			name: "multiple tokens and valid json",
			data: corev2.FixtureEntity("entity"),
			input: corev2.CheckConfig{Command: `{{ .name }}; {{ "hello" }}; {{ .entity_class }}`,
				ProxyRequests: &corev2.ProxyRequests{EntityAttributes: []string{`entity.entity_class == \"proxy\"`}},
			},
			expectedCommand: "entity; hello; host",
			expectedError:   false,
		},
		{
			name: "labels",
			data: corev2.Check{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{"foo": "bar"},
				},
			},
			input:           corev2.CheckConfig{Command: `echo {{ .labels.foo }}`},
			expectedCommand: "echo bar",
			expectedError:   false,
		},
		{
			name: "quoted strings in template",
			data: corev2.FixtureEntity("foo"),
			input: corev2.CheckConfig{
				Command: `{{ "\"hello\"" }}`,
			},
			expectedCommand: `"hello"`,
			expectedError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := TokenSubstitution(dynamic.Synthesize(tc.data), tc.input)
			testutil.CompareError(err, tc.expectedError, t)

			if !tc.expectedError {
				checkResult := corev2.CheckConfig{}
				err = json.Unmarshal(result, &checkResult)
				assert.NoError(t, err)

				assert.Equal(t, tc.expectedCommand, checkResult.Command)
			}
		})
	}
}

func TestTokenSubstitutionLabels(t *testing.T) {
	data := corev2.Check{
		ObjectMeta: corev2.ObjectMeta{
			Labels: map[string]string{"foo": "bar"},
		},
	}
	input := corev2.CheckConfig{
		ObjectMeta: corev2.ObjectMeta{
			Labels: map[string]string{
				"foo": `{{ .labels.foo }}`,
			},
		},
	}
	result, err := TokenSubstitution(dynamic.Synthesize(data), input)
	if err != nil {
		t.Fatal(err)
	}
	check := corev2.CheckConfig{}
	if err := json.Unmarshal(result, &check); err != nil {
		t.Fatal(err)
	}
	if got, want := check.Labels["foo"], "bar"; got != want {
		t.Fatalf("bad sub: got %q, want %q", got, want)
	}
}
