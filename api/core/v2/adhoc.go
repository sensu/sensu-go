package v2

import (
	"errors"
	"fmt"
	"net/url"
)

// FixtureAdhocRequest returns a testing fixture for an AdhocRequest struct.
func FixtureAdhocRequest(name string, subscriptions []string) *AdhocRequest {
	return &AdhocRequest{
		ObjectMeta:    NewObjectMeta(name, "default"),
		Subscriptions: subscriptions,
	}
}

// SetNamespace sets the namespace of the resource.
func (a *AdhocRequest) SetNamespace(namespace string) {
}

// StorePrefix returns the path prefix to this resource in the store
func (a *AdhocRequest) StorePrefix() string {
	return ""
}

// Validate returns an error if the name is not provided.
func (a *AdhocRequest) Validate() error {
	if a.Name == "" {
		return errors.New("must provide check name")
	}
	return nil
}

// URIPath is the URI path component to the adhoc request.
func (a *AdhocRequest) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/checks/%s/execute", url.PathEscape(a.Namespace), url.PathEscape(a.Name))
}

func (a *AdhocRequest) RBACName() string {
	return "checks"
}
