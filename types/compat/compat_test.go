package compat

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

func TestSetNamespace(t *testing.T) {
	tests := []struct {
		Name      string
		Resource  interface{}
		Namespace string
		Expected  string
	}{
		{
			Name:      "namespaced corev2 resource",
			Resource:  new(testCoreV2Resoruce),
			Namespace: "test",
			Expected:  "test",
		}, {
			Name:      "global corev2 resource",
			Resource:  new(testCoreV2GlobalResoruce),
			Namespace: "test",
			Expected:  "",
		}, {
			Name:      "namespaced corev3 resource",
			Resource:  new(testCoreV3Resoruce),
			Namespace: "test",
			Expected:  "test",
		}, {
			Name:      "global corev3 resource",
			Resource:  new(testCoreV3GlobalResoruce),
			Namespace: "test",
			Expected:  "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			SetNamespace(tc.Resource, tc.Namespace)
			var namespace string
			if meta := GetObjectMeta(tc.Resource); meta != nil {
				namespace = meta.Namespace
			}
			if got, want := namespace, tc.Expected; got != want {
				t.Errorf("SetNamespace had incorrect effect. got: %s, want: %s", got, want)
			}
		})
	}
}

type testCoreV3Resoruce struct {
	corev3.EntityConfig
}

type testCoreV3GlobalResoruce struct {
	corev3.EntityConfig
}

func (t *testCoreV3GlobalResoruce) IsGlobalResource() bool {
	return true
}

type testCoreV2Resoruce struct {
	corev2.Entity
}

type testCoreV2GlobalResoruce struct {
	corev2.Entity
}

func (t *testCoreV2GlobalResoruce) SetNamespace(namespace string) {}

// compile time assertions that test types
// implement expected interfaces
var _ corev3.Resource = new(testCoreV3Resoruce)
var _ corev3.Resource = new(testCoreV3GlobalResoruce)
var _ corev3.GlobalResource = new(testCoreV3GlobalResoruce)
var _ corev2.Resource = new(testCoreV2Resoruce)
var _ corev2.Resource = new(testCoreV2GlobalResoruce)
