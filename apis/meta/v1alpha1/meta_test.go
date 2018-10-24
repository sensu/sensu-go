package v1alpha1

import (
	"testing"

	"github.com/sensu/sensu-go/internal/apis/meta"
)

type TestType struct {
	meta.TypeMeta
	meta.ObjectMeta
}

func TestMetaObjectAccess(t *testing.T) {
	obj := &TestType{}
	obj.SetName("name")
	if obj.Name != "name" {
		t.Error("did not set name")
	}
	if obj.GetName() != "name" {
		t.Error("cannot access name")
	}
}

func TestGroupVersionKind(t *testing.T) {
	var obj interface{} = &TestType{
		TypeMeta: meta.TypeMeta{
			Kind:       "TestType",
			APIVersion: "apis.meta/v0",
		},
	}

	gvk, ok := obj.(meta.GroupVersionKind)
	if !ok {
		t.Error("cannot cast test type to GroupVersion")
		t.FailNow()
	}
	if gvk.GetKind() != "TestType" {
		t.Error("cannot access type")
	}
	if gvk.GetGroup() != "apis.meta" {
		t.Error("cannot access group")
	}
	if gvk.GetVersion() != "v0" {
		t.Error("cannot access version")
	}
}
