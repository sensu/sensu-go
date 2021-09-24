package v3

import (
	"errors"
	"testing"

	//nolint:staticcheck // SA1004 Replacing this will take some planning.
	"github.com/golang/protobuf/proto"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type testResource struct {
	meta *corev2.ObjectMeta
}

func (t testResource) GetMetadata() *corev2.ObjectMeta {
	if t.meta == nil {
		t.meta = &corev2.ObjectMeta{
			Name:        "foo",
			Namespace:   "default",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}
	}
	return t.meta
}

func (t testResource) SetMetadata(meta *corev2.ObjectMeta) {
	t.meta = meta //nolint:staticcheck
}

func (t testResource) StoreName() string {
	return "test"
}

func (t testResource) RBACName() string {
	return "test"
}

func (t testResource) URIPath() string {
	return "test"
}

func (t testResource) Validate() error {
	return errors.New("invalid resource")
}

func TestV3ToV2Resource(t *testing.T) {
	resource := testResource{}
	corev2Resource := V3ToV2Resource(resource)

	if got, want := corev2Resource.GetObjectMeta(), resource.GetMetadata(); !proto.Equal(&got, want) {
		t.Errorf("bad metadata: got %v, want %v", got, want)
	}

	corev2Resource.SetObjectMeta(corev2.ObjectMeta{Namespace: "bar", Name: "oof"})
	if got, want := resource.GetMetadata().Name, corev2Resource.GetObjectMeta().Name; got != want {
		t.Errorf("SetMetadata had wrong effect: got Name %s, want %s", got, want)
	}

	corev2Resource.SetNamespace("baz")
	if got, want := resource.GetMetadata().Namespace, corev2Resource.GetObjectMeta().Namespace; got != want {
		t.Errorf("SetNamespace had wrong effect: got Namespace %s, want %s", got, want)
	}

	if got, want := corev2Resource.StorePrefix(), resource.StoreName(); got != want {
		t.Errorf("bad StorePrefix: got %s, want %s", got, want)
	}

	if got, want := corev2Resource.RBACName(), resource.RBACName(); got != want {
		t.Errorf("bad RBACName: got %s, want %s", got, want)
	}

	if got, want := corev2Resource.URIPath(), resource.URIPath(); got != want {
		t.Errorf("bad URIPath: got %s, want %s", got, want)
	}

	if got, want := corev2Resource.Validate().Error(), resource.Validate().Error(); got != want {
		t.Errorf("bad Validate(): got %s, want %s", got, want)
	}
}
