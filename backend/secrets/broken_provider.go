package secrets

import (
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// BrokenProvider is a sentinel provider that can be used in place of a real
// provider, should it fail to be instantiated. Store the error from the
// failure in BrokenProvider, so the user gets the provider creation error
// every time the provider is queried.
type BrokenProvider struct {
	TypeMeta   corev2.TypeMeta
	ObjectMeta corev2.ObjectMeta
	Err        error
}

func (b *BrokenProvider) Get(id string) (string, error) {
	return "", fmt.Errorf(
		"%s/%s: %s.%s broken: cannot get secret %q: %s",
		b.ObjectMeta.Namespace,
		b.ObjectMeta.Name,
		b.TypeMeta.APIVersion,
		b.TypeMeta.Type,
		id,
		b.Err)
}

func (b *BrokenProvider) GetObjectMeta() corev2.ObjectMeta {
	return b.ObjectMeta
}

func (b *BrokenProvider) SetObjectMeta(o corev2.ObjectMeta) {
	b.ObjectMeta = o
}

func (b *BrokenProvider) SetNamespace(s string) {
	b.ObjectMeta.Namespace = s
}

func (b *BrokenProvider) StorePrefix() string {
	return ""
}

func (b *BrokenProvider) URIPath() string {
	return ""
}

func (b *BrokenProvider) RBACName() string {
	return ""
}

func (b *BrokenProvider) Validate() error {
	return nil
}
