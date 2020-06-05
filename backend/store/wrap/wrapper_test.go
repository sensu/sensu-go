package wrap

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/types"
)

func init() {
	types.RegisterTypeResolver("backend/store", testResolver)
}

func testResolver(name string) (corev2.Resource, error) {
	return corev3.V3ToV2Resource(&testResource{}), nil
}

type testResource struct {
	Metadata *corev2.ObjectMeta
}

func (t *testResource) GetMetadata() *corev2.ObjectMeta {
	return t.Metadata
}

func (t *testResource) SetMetadata(m *corev2.ObjectMeta) {
	t.Metadata = m
}

func (t *testResource) StoreSuffix() string {
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
		APIVersion: "backend/store",
	}
}

func TestWrapResourceSimple(t *testing.T) {
	resource := &testResource{
		Metadata: &corev2.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	wrapper, err := Resource(resource)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := wrapper.Metadata.APIVersion, "backend/store"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := wrapper.Metadata.Type, "testResource"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
	unwrapped, err := wrapper.Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := unwrapped, resource; !reflect.DeepEqual(got, want) {
		t.Errorf("bad resource: got %v, want %v", got, want)
	}
}
