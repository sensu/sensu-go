package token

import (
	"encoding/json"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/assert"
)

func TestSubstitution(t *testing.T) {
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
			result, err := Substitution(dynamic.Synthesize(tc.data), tc.input)
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

func TestSubstitutionLabels(t *testing.T) {
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
	result, err := Substitution(dynamic.Synthesize(data), input)
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

func TestSubstituteAsset(t *testing.T) {
	tests := []struct {
		name    string
		asset   *corev2.Asset
		entity  *corev2.Entity
		wantErr bool
		want    *corev2.Asset
	}{
		{
			name:  "A token in the URL can be substituted",
			asset: &corev2.Asset{URL: "{{ .labels.asset_url }}/asset.tar.gz"},
			entity: &corev2.Entity{ObjectMeta: corev2.ObjectMeta{
				Labels: map[string]string{"asset_url": "http://127.0.0.1"},
			}},
			want: &corev2.Asset{URL: "http://127.0.0.1/asset.tar.gz"},
		},
		{
			name:  "An asset checksum cannot be substituted",
			asset: &corev2.Asset{Sha512: "{{ .labels.sha }}"},
			entity: &corev2.Entity{ObjectMeta: corev2.ObjectMeta{
				Labels: map[string]string{"sha": "83b51af2254470edbeabf840ae556f113452133f4abbe41e0ce5e0ac37d00262a17646d38ddc23fa16f39706f3506ade902eb1b29429bb0898cfd8c5ce0b0e36"},
			}},
			want: &corev2.Asset{Sha512: "{{ .labels.sha }}"},
		},
		{
			name:    "Errors encountered while performing token substitution are returned",
			asset:   &corev2.Asset{URL: "{{ .labels.asset_url }}"},
			entity:  &corev2.Entity{},
			wantErr: true,
			want:    &corev2.Asset{URL: "{{ .labels.asset_url }}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SubstituteAsset(tt.asset, tt.entity); (err != nil) != tt.wantErr {
				t.Errorf("SubstituteAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.asset, tt.want) {
				t.Errorf("SubstituteAsset() = %#v, want %#v", tt.asset, tt.want)
			}
		})
	}
}

func TestSubstituteCheck(t *testing.T) {
	tests := []struct {
		name        string
		check       *corev2.CheckConfig
		entity      *corev2.Entity
		wantErr     bool
		wantCommand string
	}{
		{
			name:  "A token in the command can be substituted",
			check: &corev2.CheckConfig{Command: "echo {{ .labels.region }}"},
			entity: &corev2.Entity{ObjectMeta: corev2.ObjectMeta{
				Labels: map[string]string{"region": "us-west-1"},
			}},
			wantCommand: "echo us-west-1",
		},
		{
			name:        "Errors encountered while performing token substitution are returned",
			check:       &corev2.CheckConfig{Command: "echo {{ .labels.region }}"},
			entity:      &corev2.Entity{},
			wantErr:     true,
			wantCommand: "echo {{ .labels.region }}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SubstituteCheck(tt.check, tt.entity); (err != nil) != tt.wantErr {
				t.Errorf("SubstituteCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.check.Command, tt.wantCommand) {
				t.Errorf("SubstituteCheck() = %#v, want %#v", tt.check, tt.wantCommand)
			}
		})
	}
}

func TestSubstituteHook(t *testing.T) {
	tests := []struct {
		name        string
		hook        *corev2.HookConfig
		entity      *corev2.Entity
		wantErr     bool
		wantCommand string
	}{
		{
			name: "A token in the command can be substituted",
			hook: &corev2.HookConfig{Command: "echo {{ .labels.region }}"},
			entity: &corev2.Entity{ObjectMeta: corev2.ObjectMeta{
				Labels: map[string]string{"region": "us-west-1"},
			}},
			wantCommand: "echo us-west-1",
		},
		{
			name:        "Errors encountered while performing token substitution are returned",
			hook:        &corev2.HookConfig{Command: "echo {{ .labels.region }}"},
			entity:      &corev2.Entity{},
			wantErr:     true,
			wantCommand: "echo {{ .labels.region }}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SubstituteHook(tt.hook, tt.entity); (err != nil) != tt.wantErr {
				t.Errorf("SubstituteHook() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.hook.Command, tt.wantCommand) {
				t.Errorf("SubstituteHook() = %#v, want %#v", tt.hook, tt.wantCommand)
			}
		})
	}
}
