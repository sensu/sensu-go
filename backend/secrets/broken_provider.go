package secrets

import (
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// asserts that BrokenProvider implements Provider
var _ Provider = new(BrokenProvider)

// BrokenProvider is a sentinel provider that can be used in place of a real
// provider, should it fail to be instantiated. Store the error from the
// failure in BrokenProvider, so the user gets the provider creation error
// every time the provider is queried.
type BrokenProvider struct {
	TypeMeta corev2.TypeMeta
	Metadata corev2.ObjectMeta
	Err      error
}

func (b *BrokenProvider) Get(id string) (string, error) {
	return "", fmt.Errorf(
		"%s/%s: %s.%s broken: cannot get secret %q: %s",
		b.Metadata.Namespace,
		b.Metadata.Name,
		b.TypeMeta.APIVersion,
		b.TypeMeta.Type,
		id,
		b.Err)
}

func (b *BrokenProvider) GetMetadata() *corev2.ObjectMeta {
	return &b.Metadata
}

func (b *BrokenProvider) SetMetadata(o *corev2.ObjectMeta) {
	b.Metadata = *o
}

func (b *BrokenProvider) StoreName() string {
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
