package api

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

// addAuthUser finds the user in the provided context, and sets it on the
// attrs.
func addAuthUser(ctx context.Context, attrs *authorization.Attributes) error {
	// Get the claims from the request context
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return authorization.ErrNoClaims
	}

	// Add the user to our request info
	attrs.User = types.User{
		Username: claims.Subject,
		Groups:   claims.Groups,
	}

	return nil
}

// authorize is a convenience function for clients that want to test if the
// operation specified by the attributes is authorized.
func authorize(ctx context.Context, auth authorization.Authorizer, attrs *authorization.Attributes) error {
	if err := addAuthUser(ctx, attrs); err != nil {
		return err
	}
	authorized, err := auth.Authorize(ctx, attrs)
	if err != nil {
		return err
	}
	if !authorized {
		return authorization.ErrUnauthorized
	}
	return nil
}
