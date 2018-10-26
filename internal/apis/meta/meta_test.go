package meta

import (
	"testing"
)

type TestType struct {
	TypeMeta
	ObjectMeta
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
		TypeMeta: TypeMeta{
			Kind:       "TestType",
			APIVersion: "apis.meta/v0",
		},
	}

	gvk, ok := obj.(GroupVersionKind)
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
