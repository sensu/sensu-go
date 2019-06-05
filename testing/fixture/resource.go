package fixture

import (
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Resource is a resource fixture
type Resource struct {
	corev2.ObjectMeta
	Foo string
}

// GetObjectMeta ...
func (f *Resource) GetObjectMeta() corev2.ObjectMeta {
	return f.ObjectMeta
}

// SetNamespace ...
func (f *Resource) SetNamespace(namespace string) {
	f.Namespace = namespace
}

// StorePath ...
func (f *Resource) StorePath() string {
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
