package v3

// automatically generated file, do not edit!

import (
	"encoding/json"
	"net/url"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestEntityConfigStoreSuffix(t *testing.T) {
	var value EntityConfig
	got := value.StoreSuffix()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeSuffixer); ok {
		if got, want := value.StoreSuffix(), suffixer.storeSuffix(); got != want {
			t.Errorf("bad store suffix: got %s, want %s", got, want)
		}
	}
}

func TestEntityConfigRBACName(t *testing.T) {
	var value EntityConfig
	got := value.RBACName()
	if len(got) == 0 {
		t.Error("undefined rbac name")
	}
	var iface interface{} = value
	if namer, ok := iface.(rbacNamer); ok {
		if got, want := value.RBACName(), namer.rbacName(); got != want {
			t.Errorf("bad rbac name: got %s, want %s", got, want)
		}
	}
}

func TestEntityConfigURIPath(t *testing.T) {
	var value EntityConfig
	value.Metadata = &corev2.ObjectMeta{
		Namespace: "default",
		Name:      "foo",
	}
	got := value.URIPath()
	if _, err := url.Parse(got); err != nil {
		t.Error(err)
	}
	var iface interface{} = value
	if pather, ok := iface.(uriPather); ok {
		if got, want := value.URIPath(), pather.uriPath(); got != want {
			t.Errorf("bad uri path: got %s, want %s", got, want)
		}
	}
}

func TestEntityConfigValidate(t *testing.T) {
	var value EntityConfig
	if err := value.Validate(); err == nil {
		t.Errorf("expected non-nil error for nil metadata")
	}
	value.Metadata = &corev2.ObjectMeta{
		Name:        "#@$@#%@#%@#%",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	if err := value.Validate(); err == nil {
		t.Errorf("expected non-nil error for invalid metadata name")
	}
	value.Metadata.Name = "foo"
	var iface interface{} = &value
	if validator, ok := iface.(validator); ok {
		if got, want := value.Validate(), validator.validate(); got.Error() != want.Error() {
			t.Errorf("validator error: got %s, want %s", got, want)
		}
		return
	}
	if err := value.Validate(); err != nil {
		t.Error(err)
	}
}

func TestEntityConfigUnmarshalJSON(t *testing.T) {
	msg := []byte(`{"metadata": {"namespace": "default", "name": "foo"}}`)
	var value EntityConfig
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}
	var iface interface{} = &value
	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if got, want := meta.Namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := meta.Name, "foo"; got != want {
			t.Errorf("bad name: got %s, want %s", got, want)
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Error("nil annotations")
		}
	}

	// make sure labels are not accidentally zeroed out
	msg = []byte(`{"metadata": {"namespace": "default", "name": "foo", "labels": {"a": "b"}}}`)
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}

	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if got, want := len(meta.Labels), 1; got != want {
			t.Error("expected one label")
		}
	}

	// make sure annotations are not accidentally zeroed out
	msg = []byte(`{"metadata": {"namespace": "default", "name": "foo", "annotations": {"a": "b"}}}`)
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}

	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if got, want := len(meta.Annotations), 1; got != want {
			t.Error("expected one annotation")
		}
	}
}

func TestEntityStateStoreSuffix(t *testing.T) {
	var value EntityState
	got := value.StoreSuffix()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeSuffixer); ok {
		if got, want := value.StoreSuffix(), suffixer.storeSuffix(); got != want {
			t.Errorf("bad store suffix: got %s, want %s", got, want)
		}
	}
}

func TestEntityStateRBACName(t *testing.T) {
	var value EntityState
	got := value.RBACName()
	if len(got) == 0 {
		t.Error("undefined rbac name")
	}
	var iface interface{} = value
	if namer, ok := iface.(rbacNamer); ok {
		if got, want := value.RBACName(), namer.rbacName(); got != want {
			t.Errorf("bad rbac name: got %s, want %s", got, want)
		}
	}
}

func TestEntityStateURIPath(t *testing.T) {
	var value EntityState
	value.Metadata = &corev2.ObjectMeta{
		Namespace: "default",
		Name:      "foo",
	}
	got := value.URIPath()
	if _, err := url.Parse(got); err != nil {
		t.Error(err)
	}
	var iface interface{} = value
	if pather, ok := iface.(uriPather); ok {
		if got, want := value.URIPath(), pather.uriPath(); got != want {
			t.Errorf("bad uri path: got %s, want %s", got, want)
		}
	}
}

func TestEntityStateValidate(t *testing.T) {
	var value EntityState
	if err := value.Validate(); err == nil {
		t.Errorf("expected non-nil error for nil metadata")
	}
	value.Metadata = &corev2.ObjectMeta{
		Name:        "#@$@#%@#%@#%",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	if err := value.Validate(); err == nil {
		t.Errorf("expected non-nil error for invalid metadata name")
	}
	value.Metadata.Name = "foo"
	var iface interface{} = &value
	if validator, ok := iface.(validator); ok {
		if got, want := value.Validate(), validator.validate(); got.Error() != want.Error() {
			t.Errorf("validator error: got %s, want %s", got, want)
		}
		return
	}
	if err := value.Validate(); err != nil {
		t.Error(err)
	}
}

func TestEntityStateUnmarshalJSON(t *testing.T) {
	msg := []byte(`{"metadata": {"namespace": "default", "name": "foo"}}`)
	var value EntityState
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}
	var iface interface{} = &value
	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if got, want := meta.Namespace, "default"; got != want {
			t.Errorf("bad namespace: got %s, want %s", got, want)
		}
		if got, want := meta.Name, "foo"; got != want {
			t.Errorf("bad name: got %s, want %s", got, want)
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Error("nil annotations")
		}
	}

	// make sure labels are not accidentally zeroed out
	msg = []byte(`{"metadata": {"namespace": "default", "name": "foo", "labels": {"a": "b"}}}`)
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}

	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if got, want := len(meta.Labels), 1; got != want {
			t.Error("expected one label")
		}
	}

	// make sure annotations are not accidentally zeroed out
	msg = []byte(`{"metadata": {"namespace": "default", "name": "foo", "annotations": {"a": "b"}}}`)
	if err := json.Unmarshal(msg, &value); err != nil {
		t.Fatal(err)
	}

	if metaer, ok := iface.(getMetadataer); ok {
		meta := metaer.GetMetadata()
		if got, want := len(meta.Annotations), 1; got != want {
			t.Error("expected one annotation")
		}
	}
}
