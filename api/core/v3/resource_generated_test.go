package v3

// automatically generated file, do not edit!

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestEntityConfigSetMetadata(t *testing.T) {
	value := new(EntityConfig)
	var iface interface{} = value
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return
	}
	meta := &corev2.ObjectMeta{
		Name:        "snoopdogg",
		Namespace:   "lbc",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	value.SetMetadata(meta)
	if got, want := metaer.GetMetadata(), meta; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad metadata: got %v, want %v", got, want)
	}
}

func TestEntityConfigStoreName(t *testing.T) {
	var value EntityConfig
	got := value.StoreName()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeNamer); ok {
		if got, want := value.StoreName(), suffixer.storeName(); got != want {
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

func TestEntityConfigGetTypeMeta(t *testing.T) {
	var value EntityConfig
	meta := value.GetTypeMeta()
	if got, want := meta.APIVersion, "core/v3"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := meta.Type, "EntityConfig"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
}

func TestEntityStateSetMetadata(t *testing.T) {
	value := new(EntityState)
	var iface interface{} = value
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return
	}
	meta := &corev2.ObjectMeta{
		Name:        "snoopdogg",
		Namespace:   "lbc",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	value.SetMetadata(meta)
	if got, want := metaer.GetMetadata(), meta; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad metadata: got %v, want %v", got, want)
	}
}

func TestEntityStateStoreName(t *testing.T) {
	var value EntityState
	got := value.StoreName()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeNamer); ok {
		if got, want := value.StoreName(), suffixer.storeName(); got != want {
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

func TestEntityStateGetTypeMeta(t *testing.T) {
	var value EntityState
	meta := value.GetTypeMeta()
	if got, want := meta.APIVersion, "core/v3"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := meta.Type, "EntityState"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
}

func TestPipelineSetMetadata(t *testing.T) {
	value := new(Pipeline)
	var iface interface{} = value
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return
	}
	meta := &corev2.ObjectMeta{
		Name:        "snoopdogg",
		Namespace:   "lbc",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	value.SetMetadata(meta)
	if got, want := metaer.GetMetadata(), meta; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad metadata: got %v, want %v", got, want)
	}
}

func TestPipelineStoreName(t *testing.T) {
	var value Pipeline
	got := value.StoreName()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeNamer); ok {
		if got, want := value.StoreName(), suffixer.storeName(); got != want {
			t.Errorf("bad store suffix: got %s, want %s", got, want)
		}
	}
}

func TestPipelineRBACName(t *testing.T) {
	var value Pipeline
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

func TestPipelineURIPath(t *testing.T) {
	var value Pipeline
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

func TestPipelineValidate(t *testing.T) {
	var value Pipeline
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

func TestPipelineUnmarshalJSON(t *testing.T) {
	msg := []byte(`{"metadata": {"namespace": "default", "name": "foo"}}`)
	var value Pipeline
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

func TestPipelineGetTypeMeta(t *testing.T) {
	var value Pipeline
	meta := value.GetTypeMeta()
	if got, want := meta.APIVersion, "core/v3"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := meta.Type, "Pipeline"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
}

func TestResourceTemplateSetMetadata(t *testing.T) {
	value := new(ResourceTemplate)
	var iface interface{} = value
	metaer, ok := iface.(getMetadataer)
	if !ok {
		return
	}
	meta := &corev2.ObjectMeta{
		Name:        "snoopdogg",
		Namespace:   "lbc",
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	value.SetMetadata(meta)
	if got, want := metaer.GetMetadata(), meta; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad metadata: got %v, want %v", got, want)
	}
}

func TestResourceTemplateStoreName(t *testing.T) {
	var value ResourceTemplate
	got := value.StoreName()
	if len(got) == 0 {
		t.Error("undefined store suffix")
	}
	var iface interface{} = value
	if suffixer, ok := iface.(storeNamer); ok {
		if got, want := value.StoreName(), suffixer.storeName(); got != want {
			t.Errorf("bad store suffix: got %s, want %s", got, want)
		}
	}
}

func TestResourceTemplateRBACName(t *testing.T) {
	var value ResourceTemplate
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

func TestResourceTemplateURIPath(t *testing.T) {
	var value ResourceTemplate
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

func TestResourceTemplateValidate(t *testing.T) {
	var value ResourceTemplate
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

func TestResourceTemplateUnmarshalJSON(t *testing.T) {
	msg := []byte(`{"metadata": {"namespace": "default", "name": "foo"}}`)
	var value ResourceTemplate
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

func TestResourceTemplateGetTypeMeta(t *testing.T) {
	var value ResourceTemplate
	meta := value.GetTypeMeta()
	if got, want := meta.APIVersion, "core/v3"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := meta.Type, "ResourceTemplate"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
}

func TestResourceUniqueness(t *testing.T) {
	types := make(map[reflect.Type]GeneratedType)
	for _, v := range typeMap {
		types[reflect.TypeOf(v)] = v.(GeneratedType)
	}
	if got, want := len(types), len(typeMap)/2; got != want {
		t.Fatalf("bad number of types: got %d, want %d", got, want)
	}
	rbacNames := make(map[string]bool)
	for _, v := range types {
		if name := v.RBACName(); rbacNames[name] {
			t.Errorf("duplicate rbac name: %s", name)
		} else {
			rbacNames[name] = true
		}
	}
	storeNames := make(map[string]bool)
	for _, v := range types {
		if name := v.StoreName(); storeNames[name] {
			t.Errorf("duplicate store suffix: %s", name)
		} else {
			storeNames[name] = true
		}
	}
}
