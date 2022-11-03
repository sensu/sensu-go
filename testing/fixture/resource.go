package fixture

import (
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// GetObjectMeta ...
func (f *Resource) GetObjectMeta() corev2.ObjectMeta {
	return f.ObjectMeta
}

// SetObjectMeta ...
func (f *Resource) SetObjectMeta(meta corev2.ObjectMeta) {
	f.ObjectMeta = meta
}

// SetNamespace ...
func (f *Resource) SetNamespace(namespace string) {
	f.Namespace = namespace
}

// StorePrefix ...
func (f *Resource) StorePrefix() string {
	return "resource"
}

// URIPath ...
func (f *Resource) URIPath() string {
	return fmt.Sprintf("/resource/%s", f.Name)
}

// Validate ...
func (f *Resource) Validate() error {
	return nil
}

func (*Resource) RBACName() string {
	return "resource"
}
