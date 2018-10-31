package types

import (
	"fmt"
	"net/url"
)

const (
	// NamespaceTypeAll matches all actions
	NamespaceTypeAll = "*"
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
	return fmt.Sprintf("/rbac/namespaces/%s", url.PathEscape(n.Name))
}
