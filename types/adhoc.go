package types

import (
	"errors"
	"fmt"
	"net/url"
)

// Validate returns an error if the name is not provided.
func (a *AdhocRequest) Validate() error {
	if a.Name == "" {
		return errors.New("must provide check name")
	}
	return nil
}

// FixtureAdhocRequest returns a testing fixture for an AdhocRequest struct.
func FixtureAdhocRequest(name string, subscriptions []string) *AdhocRequest {
	return &AdhocRequest{
		Name:          name,
		Subscriptions: subscriptions,
	}
}

// URIPath is the URI path component to the adhoc request.
func (a *AdhocRequest) URIPath() string {
	return fmt.Sprintf("/checks/%s/execute", url.PathEscape(a.Name))
}
