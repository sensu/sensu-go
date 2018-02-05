package types

import "errors"

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
