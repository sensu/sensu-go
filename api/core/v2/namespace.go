package v2

import (
	"fmt"
	"net/url"
)

const (
	// NamespaceTypeAll represents an empty namespace, used to represent a request
	// across all namespaces
	NamespaceTypeAll = ""
)

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

// URIPath returns the path component of a Namespace URI.
func (n *Namespace) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s", url.PathEscape(n.Name))
}

// GetObjectMeta only exists here to fulfil the requirements of Resource
func (n *Namespace) GetObjectMeta() ObjectMeta {
	return ObjectMeta{}
}

// NamespaceFields returns a set of fields that represent that resource
func NamespaceFields(r Resource) map[string]string {
	resource := r.(*Namespace)
	return map[string]string{
		"namespace.name": resource.Name,
	}
}
