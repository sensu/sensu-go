package authorization

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// Authorizer determines whether a request is authorized using the Attributes
// passed. It returns true if the request should be authorized, along with any
// error encountered
type Authorizer interface {
	Authorize(attrs *Attributes) (bool, error)
}

// Attributes represents all the information required by an authorizer to make
// an authorization decision
type Attributes struct {
	APIGroup     string
	APIVersion   string
	Namespace    string
	Resource     string
	ResourceName string
	User         types.User
	Verb         string
}

// GetAttributes returns the authorization attributes stored in the given
// context
func GetAttributes(ctx context.Context) *Attributes {
	if value := ctx.Value(types.AuthorizationAttributesKey); value != nil {
		return value.(*Attributes)
	}
	return nil
}

// SetAttributes stores the given attributes within the provided context
func SetAttributes(ctx context.Context, attrs *Attributes) context.Context {
	return context.WithValue(ctx, types.AuthorizationAttributesKey, attrs)
}
