package v2

import (
	"fmt"
	"net/url"
	"path"
)

const (
	// NamespaceTypeAll represents an empty namespace, used to represent a request
	// across all namespaces
	NamespaceTypeAll = ""

	// NamespacesResource is the name of this resource type
	NamespacesResource = "namespaces"
)

// StorePrefix returns the path prefix to this resource in the store
func (n *Namespace) StorePrefix() string {
	return NamespacesResource
}

// URIPath returns the path component of a Namespace URI.
func (n *Namespace) URIPath() string {
	return path.Join(URLPrefix, "namespaces", url.PathEscape(n.Name))
}

// Validate returns an error if the namespace does not pass validation tests
func (n *Namespace) Validate() error {
	if err := ValidateName(n.Name); err != nil {
		return fmt.Errorf("namespace name %s", err)
	}

	return nil
}

// FixtureNamespace returns a mocked namespace
func FixtureNamespace(name string) *Namespace {
	return &Namespace{
		Name: name,
	}
}

// GetObjectMeta only exists here to fulfil the requirements of Resource
func (n *Namespace) GetObjectMeta() ObjectMeta {
	return ObjectMeta{Name: n.Name}
}

// NamespaceFields returns a set of fields that represent that resource
func NamespaceFields(r Resource) map[string]string {
	resource := r.(*Namespace)
	return map[string]string{
		"namespace.name": resource.Name,
	}
}

// SetNamespace sets the namespace of the resource.
func (n *Namespace) SetNamespace(namespace string) {
}

// SetObjectMeta only exists here to fulfil the requirements of Resource
func (n *Namespace) SetObjectMeta(meta ObjectMeta) {
	n.Name = meta.Name
}

func (n *Namespace) RBACName() string {
	return "namespaces"
}
