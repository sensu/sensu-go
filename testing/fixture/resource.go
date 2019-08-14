package fixture

import (
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// GetObjectMeta ...
func (f *Resource) GetObjectMeta() corev2.ObjectMeta {
	return f.ObjectMeta
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
