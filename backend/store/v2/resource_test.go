package v2

import (
	"testing"

	corev2 "github.com/sensu/core/v2"
)

type testResource struct {
	Metadata *corev2.ObjectMeta
}

func (t *testResource) GetMetadata() *corev2.ObjectMeta {
	return t.Metadata
}

func (t *testResource) SetMetadata(m *corev2.ObjectMeta) {
	t.Metadata = m
}

func (t *testResource) StoreName() string {
	return "testresource"
}

func (t *testResource) RBACName() string {
	return "testresource"
}

func (t *testResource) URIPath() string {
	return "api/backend/store/namespaces/default/testresource/test"
}

func (t *testResource) Validate() error {
	return nil
}

func (t *testResource) GetTypeMeta() corev2.TypeMeta {
	return corev2.TypeMeta{
		Type:       "testResource",
		APIVersion: "store/wrap_test",
	}
}

func fixtureTestResource(name string) *testResource {
	return &testResource{
		Metadata: &corev2.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

func TestNewResourceRequest(t *testing.T) {
	req := NewResourceRequest(corev2.TypeMeta{Type: "CheckConfig", APIVersion: "core/v2"}, "default", "foo", "checks")
	if got, want := req.Namespace, "default"; got != want {
		t.Errorf("bad namespace: got %s, want %s", got, want)
	}
	if got, want := req.Name, "foo"; got != want {
		t.Errorf("bad name: got %s, want %s", got, want)
	}
	if got, want := req.StoreName, "checks"; got != want {
		t.Errorf("bad store name: got %s, want %s", got, want)
	}
}

func TestNewResourceRequestFromResource(t *testing.T) {
	resource := fixtureTestResource("foo")
	req := NewResourceRequestFromResource(resource)
	if got, want := req.Name, resource.GetMetadata().GetName(); got != want {
		t.Errorf("bad name: got %s, want %s", got, want)
	}
	if got, want := req.Namespace, resource.GetMetadata().GetNamespace(); got != want {
		t.Errorf("bad namespace: got %s, want %s", got, want)
	}
	if got, want := req.StoreName, resource.StoreName(); got != want {
		t.Errorf("bad store name: got %s, want %s", got, want)
	}
	if got, want := req.Type, "testResource"; got != want {
		t.Errorf("bad type metadata: got %v, want %v", got, want)
	}
}
