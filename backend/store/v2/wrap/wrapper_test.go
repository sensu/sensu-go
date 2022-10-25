package wrap_test

import (
	"encoding/json"
	fmt "fmt"
	"testing"

	//nolint:staticcheck // SA1004 Replacing this will take some planning.
	"github.com/golang/protobuf/proto"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/types"
)

func init() {
	types.RegisterResolver("wrap_test/v2", testResolver)
	types.RegisterResolver("v2/wrap_test", testResolver)
}

func testResolver(name string) (interface{}, error) {
	switch name {
	case "testResource":
		return &testResource{}, nil
	case "testResource2":
		return &testResource2{}, nil
	default:
		return nil, fmt.Errorf("invalid resource: %s", name)
	}
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
		APIVersion: "wrap_test/v2",
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

func TestWrapResourceSimple(t *testing.T) {
	resource := fixtureTestResource("test")
	wrapper, err := wrap.Resource(resource)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := wrapper.TypeMeta.APIVersion, "wrap_test/v2"; got != want {
		t.Errorf("bad api version: got %s, want %s", got, want)
	}
	if got, want := wrapper.TypeMeta.Type, "testResource"; got != want {
		t.Errorf("bad type: got %s, want %s", got, want)
	}
	unwrapped, err := wrapper.Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := unwrapped.GetMetadata(), resource.GetMetadata(); !proto.Equal(got, want) {
		t.Errorf("bad resource: got %#v, want %#v", got, want)
	}
}

func TestWrapResourceOptions(t *testing.T) {
	tests := []struct {
		Name     string
		Resource corev3.Resource
		Options  []wrap.Option
		ExpErr   bool
		TestHook func(testing.TB, *wrap.Wrapper, corev3.Resource)
	}{
		{
			Name:     "force protobuf on resource that is not a proto.Message",
			Resource: fixtureTestResource("forceproto"),
			Options:  []wrap.Option{wrap.EncodeProtobuf},
			ExpErr:   true,
		},
		{
			Name:     "disable compression",
			Resource: fixtureTestResource("disablecompression"),
			Options:  []wrap.Option{wrap.CompressNone},
			TestHook: func(t testing.TB, w *wrap.Wrapper, r corev3.Resource) {
				t.Helper()
				var msg *json.RawMessage
				if err := json.Unmarshal(w.Value, &msg); err != nil {
					t.Error(err)
				}
			},
		},
		{
			Name:     "protobuf",
			Resource: corev3.FixtureEntityState("estate"),
			TestHook: func(t testing.TB, w *wrap.Wrapper, r corev3.Resource) {
				t.Helper()
				var state corev3.EntityState
				if err := proto.Unmarshal(w.Value, &state); err != nil {
					t.Fatal(err)
				}
			},
			Options: []wrap.Option{wrap.CompressNone},
		},
		{
			Name:     "no type meta",
			Resource: fixtureTestResource2("tr2"),
		},
		{
			Name:     "force json encoding on proto messages",
			Resource: corev3.FixtureEntityState("estate"),
			Options:  []wrap.Option{wrap.CompressNone, wrap.EncodeJSON},
			TestHook: func(t testing.TB, w *wrap.Wrapper, r corev3.Resource) {
				t.Helper()
				var msg *json.RawMessage
				if err := json.Unmarshal(w.Value, &msg); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrapper, err := wrap.Resource(test.Resource, test.Options...)
			if err != nil {
				if !test.ExpErr {
					t.Fatal(err)
				} else {
					return
				}
			} else {
				if test.ExpErr {
					t.Fatal(err)
				}
			}
			if test.TestHook != nil {
				test.TestHook(t, wrapper, test.Resource)
			}
			resource, err := wrapper.Unwrap()
			if err != nil {
				t.Fatal(err)
			}
			if got, want := resource.GetMetadata(), test.Resource.GetMetadata(); !proto.Equal(got, want) {
				t.Errorf("bad resource: got %v, want %v", got, want)
			}
		})
	}
}

type testResource2 struct {
	Metadata *corev2.ObjectMeta
}

func (t *testResource2) GetMetadata() *corev2.ObjectMeta {
	return t.Metadata
}

func (t *testResource2) SetMetadata(m *corev2.ObjectMeta) {
	t.Metadata = m
}

func (t *testResource2) StoreName() string {
	return "testresource2"
}

func (t *testResource2) RBACName() string {
	return "testresource2"
}

func (t *testResource2) URIPath() string {
	return "api/backend/store/namespaces/default/testresource2/test"
}

func (t *testResource2) Validate() error {
	return nil
}

func (t *testResource2) GetTypeMeta() corev2.TypeMeta {
	return corev2.TypeMeta{
		Type:       "testResource2",
		APIVersion: "wrap_test/v2",
	}
}

func fixtureTestResource2(name string) *testResource2 {
	return &testResource2{
		Metadata: &corev2.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}
