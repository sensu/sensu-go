package edit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"gopkg.in/yaml.v2"
)

type testConfig struct {
	namespace string
	format    string
}

func (t testConfig) Namespace() string {
	return t.namespace
}

func (t testConfig) Format() string {
	return t.format
}

type testClient struct {
	err error
}

func (t testClient) Get(_ string, val interface{}) error {
	if t.err != nil {
		return t.err
	}
	switch v := val.(type) {
	case *corev3.Namespace:
		*v = *(corev3.FixtureNamespace("default"))
	case *corev2.ClusterRole:
		*v = *(corev2.FixtureClusterRole("default"))
	case *corev2.ClusterRoleBinding:
		*v = *(corev2.FixtureClusterRoleBinding("default"))
	case *corev2.User:
		*v = *(corev2.FixtureUser("default"))
	case *corev2.TessenConfig:
	case *corev2.Asset:
		*v = *(corev2.FixtureAsset("default"))
	case *corev2.CheckConfig:
		*v = *(corev2.FixtureCheckConfig("default"))
	case *corev2.Entity:
		*v = *(corev2.FixtureEntity("default"))
	case *corev2.Event:
		*v = *(corev2.FixtureEvent("default", "default"))
	case *corev2.EventFilter:
		*v = *(corev2.FixtureEventFilter("default"))
	case *corev2.Handler:
		*v = *(corev2.FixtureHandler("default"))
	case *corev2.Hook:
		*v = *(corev2.FixtureHook("default"))
	case *corev2.Mutator:
		*v = *(corev2.FixtureMutator("default"))
	case *corev2.Role:
		*v = *(corev2.FixtureRole("default", "default"))
	case *corev2.RoleBinding:
		*v = *(corev2.FixtureRoleBinding("default", "default"))
	case *corev2.Silenced:
		*v = *(corev2.FixtureSilenced("default"))
	}
	return nil
}

type um int

func (t um) String() string {
	switch t {
	case 0:
		return "json"
	case 1:
		return "yaml"
	}
	panic("!")
}

func testName(i um, name string) string {
	return fmt.Sprintf("%s-%s", name, i)
}

func TestDumpResource(t *testing.T) {
	tests := []struct {
		Type string
		Key  []string
		Err  bool
	}{
		{
			Type: "namespace",
			Key:  []string{"default"},
		},
		{
			Type: "cluster_role",
			Key:  []string{"default"},
		},
		{
			Type: "cluster_role_binding",
			Key:  []string{"default"},
		},
		{
			Type: "user",
			Key:  []string{"default"},
		},
		{
			Type: "asset",
			Key:  []string{"default"},
		},
		{
			Type: "check_config",
			Key:  []string{"default"},
		},
		{
			Type: "check",
			Key:  []string{"default"},
		},
		{
			Type: "entity",
			Key:  []string{"default"},
		},
		{
			Type: "event",
			Key:  []string{"default", "default"},
		},
		{
			Type: "event",
			Key:  []string{"default"},
			Err:  true,
		},
		{
			Type: "event_filter",
			Key:  []string{"default"},
		},
		{
			Type: "handler",
			Key:  []string{"default"},
		},
		{
			Type: "mutator",
			Key:  []string{"default"},
		},
		{
			Type: "role",
			Key:  []string{"default"},
		},
		{
			Type: "role_binding",
			Key:  []string{"default"},
		},
		{
			Type: "silenced",
			Key:  []string{"default"},
		},
		{
			Type: "silenced",
			Err:  true,
		},
		{
			Type: "invalid",
			Err:  true,
		},
	}
	for _, test := range tests {
		unmarshalers := []func([]byte, interface{}) error{
			json.Unmarshal,
			yaml.Unmarshal,
		}
		for i, unmarshal := range unmarshalers {
			t.Run(testName(um(i), test.Type), func(t *testing.T) {
				cfg := testConfig{
					namespace: "default",
					format:    "json",
				}
				client := testClient{}
				buf := new(bytes.Buffer)
				err := dumpResource(client, cfg, test.Type, test.Key, buf)
				if err != nil && !test.Err {
					t.Error(err)
				}
				if err == nil && test.Err {
					t.Error("expected non-nil error")
				}
				if test.Err {
					return
				}
				var m map[string]interface{}
				if err := unmarshal(buf.Bytes(), &m); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestDumpBlank(t *testing.T) {
	tests := []struct {
		Type string
		Err  bool
	}{
		{
			Type: "namespace",
		},
		{
			Type: "cluster_role",
		},
		{
			Type: "cluster_role_binding",
		},
		{
			Type: "user",
		},
		{
			Type: "asset",
		},
		{
			Type: "check_config",
		},
		{
			Type: "check",
		},
		{
			Type: "entity",
		},
		{
			Type: "event",
		},
		{
			Type: "event_filter",
		},
		{
			Type: "handler",
		},
		{
			Type: "mutator",
		},
		{
			Type: "role",
		},
		{
			Type: "role_binding",
		},
		{
			Type: "silenced",
		},
		{
			Type: "invalid",
			Err:  true,
		},
	}
	for _, test := range tests {
		unmarshalers := []func([]byte, interface{}) error{
			json.Unmarshal,
			yaml.Unmarshal,
		}
		for i, unmarshal := range unmarshalers {
			t.Run(testName(um(i), test.Type), func(t *testing.T) {
				cfg := testConfig{
					namespace: "default",
					format:    "json",
				}
				buf := new(bytes.Buffer)
				err := dumpBlank(cfg, test.Type, buf)
				if err != nil && !test.Err {
					t.Error(err)
				}
				if err == nil && test.Err {
					t.Error("expected non-nil error")
				}
				if test.Err {
					return
				}
				var m map[string]interface{}
				if err := unmarshal(buf.Bytes(), &m); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		Input  string
		Output []string
	}{
		{
			Input:  "foo bar baz",
			Output: []string{"foo", "bar", "baz"},
		},
		{
			Input:  "foo\tbar",
			Output: []string{"foo", "bar"},
		},
		{
			Input:  "foo\tbar\n\n\n\nbaz",
			Output: []string{"foo", "bar", "baz"},
		},
	}
	for _, test := range tests {
		if got, want := parseCommand(test.Input), test.Output; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad result: got %v, want %v", got, want)
		}
	}
}
