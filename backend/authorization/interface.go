package authorization

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/types"
)

var ErrUnauthorized = errors.New("request unauthorized")
var ErrNoClaims = errors.New("no claims found in the request context")

// Authorizer determines whether a request is authorized using the Attributes
// passed. It returns true if the request should be authorized, along with any
// error encountered
type Authorizer interface {
	Authorize(ctx context.Context, attrs *Attributes) (bool, error)
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

// AttributesKey is a convenience type for storing an attributes-like value
// as a map key.
type AttributesKey struct {
	APIGroup     string
	APIVersion   string
	Namespace    string
	Resource     string
	ResourceName string
	UserName     string
	Verb         string
}

func (a Attributes) Key() AttributesKey {
	return AttributesKey{
		APIGroup:     a.APIGroup,
		APIVersion:   a.APIVersion,
		Namespace:    a.Namespace,
		Resource:     a.Resource,
		ResourceName: a.ResourceName,
		UserName:     a.User.Username,
		Verb:         a.Verb,
	}
}
