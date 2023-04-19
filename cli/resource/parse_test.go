package resource

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/go-test/deep"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/util/compat"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		resource      *types.Wrapper
		namespace     string
		wantNamespace string
	}{
		{
			name: "a namespaced resource with a configured namespace should not be modified",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				},
			},
			namespace:     "dev",
			wantNamespace: "default",
		},
		{
			name: "a namespaced resource without a configured namespace should use the provided namespace",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "dev",
		},
		{
			name: "a global resource should not have a namespace configured",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				Value: &corev2.ClusterRole{
					ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := []*types.Wrapper{tt.resource}
			_ = Validate(resources, tt.namespace)

			if tt.resource.ObjectMeta.Namespace != tt.wantNamespace {
				t.Errorf("Validate() wrapper namespace = %q, want namespace %q", tt.resource.ObjectMeta.Namespace, tt.wantNamespace)
			}
			if tt.resource.Value != nil && compat.GetObjectMeta(tt.resource.Value).Namespace != tt.wantNamespace {
				t.Errorf("Validate() wrapper's resource namespace = %q, want namespace %q", compat.GetObjectMeta(tt.resource.Value).Namespace, tt.wantNamespace)
			}
		})
	}
}

func TestValidateStderr(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ch := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		ch <- buf.String()
	}()

	resources := []*types.Wrapper{&types.Wrapper{
		ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
	}}
	_ = Validate(resources, "default")

	// Reset stderr
	w.Close()
	os.Stderr = oldStderr

	errMsg := <-ch
	errMsg = strings.TrimSpace(errMsg)
	wantErr := `error validating resource #0 with name "check-cpu" and namespace "default": resource is nil`
	if errMsg != wantErr {
		t.Errorf("Validate() err = %s, want %s", errMsg, wantErr)
	}
}

func TestParse(t *testing.T) {
	const (
		jsonUnix                  = "{\n  \"type\": \"EventFilter\",\n  \"api_version\": \"core/v2\",\n  \"metadata\": {\n    \"name\": \"filter_minimum\",\n    \"namespace\": \"default\"\n  },\n  \"spec\": {\n    \"action\": \"allow\",\n    \"expressions\": [\n      \"event.check.occurrences == 1\"\n    ]\n  }\n}"
		jsonWindows               = "{\r\n  \"type\": \"EventFilter\",\r\n  \"api_version\": \"core/v2\",\r\n  \"metadata\": {\r\n    \"name\": \"filter_minimum\",\r\n    \"namespace\": \"default\"\r\n  },\r\n  \"spec\": {\r\n    \"action\": \"allow\",\r\n    \"expressions\": [\r\n      \"event.check.occurrences == 1\"\r\n    ]\r\n  }\r\n}"
		jsonError                 = "{\n  {\"type\": \"EventFilter\",\n  \"api_version\": \"core/v2\",\n  \"metadata\": {\n    \"name\": \"filter_minimum\",\n    \"namespace\": \"default\"\n  },\n  \"spec\": {\n    \"action\": \"allow\",\n    \"expressions\": [\n      \"event.check.occurrences == 1\"\n    ]\n  }\n}"
		yamlUnixSingle            = "api_version: core/v2\ntype: Handler\nmetadata:\n  namespace: default\n  name: email\nspec:\n  type: pipe\n  command: sensu-email-handler \n    -u USERNAME -p PASSWORD\n  timeout: 10\n  filters:\n  - is_incident\n  - not_silenced\n  - state_change_only\n  runtime_assets:\n  - email-handler\n"
		yamlWindowsSingle         = "api_version: core/v2\r\ntype: Handler\r\nmetadata:\r\n  namespace: default\r\n  name: email\r\nspec:\r\n  type: pipe\r\n  command: sensu-email-handler \r\n    -u USERNAME -p PASSWORD\r\n  timeout: 10\r\n  filters:\r\n  - is_incident\r\n  - not_silenced\r\n  - state_change_only\r\n  runtime_assets:\r\n  - email-handler\r\n"
		yamlUnixSinglePrefixed    = "---\napi_version: core/v2\ntype: Handler\nmetadata:\n  namespace: default\n  name: email\nspec:\n  type: pipe\n  command: sensu-email-handler \n    -u USERNAME -p PASSWORD\n  timeout: 10\n  filters:\n  - is_incident\n  - not_silenced\n  - state_change_only\n  runtime_assets:\n  - email-handler\n"
		yamlWindowsSinglePrefixed = "---\r\napi_version: core/v2\r\ntype: Handler\r\nmetadata:\r\n  namespace: default\r\n  name: email\r\nspec:\r\n  type: pipe\r\n  command: sensu-email-handler \r\n    -u USERNAME -p PASSWORD\r\n  timeout: 10\r\n  filters:\r\n  - is_incident\r\n  - not_silenced\r\n  - state_change_only\r\n  runtime_assets:\r\n  - email-handler\r\n"
		yamlUnixMulti             = "type: CheckConfig\napi_version: core/v2\nmetadata:\n  name: foo\nspec:\n  command: echo foo\n  interval: 100\n--- # comment\napi_version: core/v2\ntype: Handler\nmetadata:\n  namespace: default\n  name: email\nspec:\n  type: pipe\n  command: sensu-email-handler \n    -u USERNAME -p PASSWORD\n  timeout: 10\n  filters:\n  - is_incident\n  - not_silenced\n  - state_change_only\n  runtime_assets:\n  - email-handler\n"
		yamlWindowsMulti          = "type: CheckConfig\r\napi_version: core/v2\r\nmetadata:\r\n  name: foo\r\nspec:\r\n  command: echo foo\r\n  interval: 100\r\n--- # comment\r\napi_version: core/v2\r\ntype: Handler\r\nmetadata:\r\n  namespace: default\r\n  name: email\r\nspec:\r\n  type: pipe\r\n  command: sensu-email-handler \r\n    -u USERNAME -p PASSWORD\r\n  timeout: 10\r\n  filters:\r\n  - is_incident\r\n  - not_silenced\r\n  - state_change_only\r\n  runtime_assets:\r\n  - email-handler\r\n"
		yamlUnixMultiPrefixed     = "---\ntype: CheckConfig\napi_version: core/v2\nmetadata:\n  name: foo\nspec:\n  command: echo foo\n  interval: 100\n--- # comment\napi_version: core/v2\ntype: Handler\nmetadata:\n  namespace: default\n  name: email\nspec:\n  type: pipe\n  command: sensu-email-handler \n    -u USERNAME -p PASSWORD\n  timeout: 10\n  filters:\n  - is_incident\n  - not_silenced\n  - state_change_only\n  runtime_assets:\n  - email-handler\n"
		yamlWindowsMultiPrefixed  = "---\ntype: CheckConfig\r\napi_version: core/v2\r\nmetadata:\r\n  name: foo\r\nspec:\r\n  command: echo foo\r\n  interval: 100\r\n--- # comment\r\napi_version: core/v2\r\ntype: Handler\r\nmetadata:\r\n  namespace: default\r\n  name: email\r\nspec:\r\n  type: pipe\r\n  command: sensu-email-handler \r\n    -u USERNAME -p PASSWORD\r\n  timeout: 10\r\n  filters:\r\n  - is_incident\r\n  - not_silenced\r\n  - state_change_only\r\n  runtime_assets:\r\n  - email-handler\r\n"
		yamlError                 = "%$^apiVersion: core/v2\ntype: Handler\nmetadata:\n  namespace: default\n  name: email\nspec:\n  type: pipe\n  command: sensu-email-handler \n    -u USERNAME -p PASSWORD\n  timeout: 10\n  filters:\n  - is_incident\n  - not_silenced\n  - state_change_only\n  runtime_assets:\n  - email-handler\n"
	)

	checkConfigWrapper := &types.Wrapper{
		ObjectMeta: corev2.ObjectMeta{
			Name: "foo",
		},
		TypeMeta: corev2.TypeMeta{
			Type:       "CheckConfig",
			APIVersion: "core/v2",
		},
		Value: &corev2.CheckConfig{
			ObjectMeta: corev2.ObjectMeta{
				Name:        "foo",
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Command:  "echo foo",
			Interval: 100,
		},
	}

	handlerWrapper := &types.Wrapper{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "email",
		},
		TypeMeta: corev2.TypeMeta{
			Type:       "Handler",
			APIVersion: "core/v2",
		},
		Value: &corev2.Handler{
			ObjectMeta: corev2.ObjectMeta{
				Namespace:   "default",
				Name:        "email",
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Type:          "pipe",
			Command:       "sensu-email-handler -u USERNAME -p PASSWORD",
			Timeout:       10,
			Filters:       []string{"is_incident", "not_silenced", "state_change_only"},
			RuntimeAssets: []string{"email-handler"},
		},
	}

	eventFilterWrapper := &types.Wrapper{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "filter_minimum",
		},
		TypeMeta: corev2.TypeMeta{
			Type:       "EventFilter",
			APIVersion: "core/v2",
		},
		Value: &corev2.EventFilter{
			ObjectMeta: corev2.ObjectMeta{
				Namespace:   "default",
				Name:        "filter_minimum",
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Action:      "allow",
			Expressions: []string{"event.check.occurrences == 1"},
		},
	}

	tests := []struct {
		name        string
		fileContent string
		want        []*types.Wrapper
		wantErr     bool
		wantErrMsg  string
	}{
		{
			name:        "should parse a single unix formatted json resource",
			fileContent: jsonUnix,
			want:        []*types.Wrapper{eventFilterWrapper},
		},
		{
			name:        "should parse a single windows formatted json resource",
			fileContent: jsonWindows,
			want:        []*types.Wrapper{eventFilterWrapper},
		},
		{
			name:        "should parse a single unix formatted yaml resource",
			fileContent: yamlUnixSingle,
			want:        []*types.Wrapper{handlerWrapper},
		},
		{
			name:        "should parse a single windows formatted yaml resource",
			fileContent: yamlWindowsSingle,
			want:        []*types.Wrapper{handlerWrapper},
		},
		{
			name:        "should parse a single unix formatted yaml resource prefixed with ---",
			fileContent: yamlUnixSinglePrefixed,
			want:        []*types.Wrapper{handlerWrapper},
		},
		{
			name:        "should parse a single windows formatted yaml resource prefixed with ---",
			fileContent: yamlWindowsSinglePrefixed,
			want:        []*types.Wrapper{handlerWrapper},
		},
		{
			name:        "should parse multiple unix formatted yaml resources",
			fileContent: yamlUnixMulti,
			want:        []*types.Wrapper{checkConfigWrapper, handlerWrapper},
		},
		{
			name:        "should parse multiple windows formatted yaml resources",
			fileContent: yamlWindowsMulti,
			want:        []*types.Wrapper{checkConfigWrapper, handlerWrapper},
		},
		{
			name:        "should parse multiple unix formatted yaml resources prefixed with ---",
			fileContent: yamlUnixMultiPrefixed,
			want:        []*types.Wrapper{checkConfigWrapper, handlerWrapper},
		},
		{
			name:        "should parse multiple windows formatted yaml resources prefixed with ---",
			fileContent: yamlWindowsMultiPrefixed,
			want:        []*types.Wrapper{checkConfigWrapper, handlerWrapper},
		},
		{
			name:        "should return an error when parsing a badly formatted json file",
			fileContent: jsonError,
			wantErr:     true,
			wantErrMsg:  "too many errors",
		},
		{
			name:        "should return an error when parsing a badly formatted yaml file",
			fileContent: yamlError,
			wantErr:     true,
			wantErrMsg:  "error parsing resources: yaml: could not find expected directive name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stringReader := strings.NewReader(tt.fileContent)
			got, err := Parse(stringReader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErrMsg != err.Error() {
				t.Errorf("Parse() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Parse() got differs from want: %v", diff)
			}
		})
	}
}
