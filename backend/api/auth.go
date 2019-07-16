package api

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

func addAuthUser(ctx context.Context, attrs *authorization.Attributes) error {
	// Get the claims from the request context
	claims := jwt.GetClaimsFromContext(ctx)
	if claims == nil {
		return errors.New("no claims found in the request context")
	}

	// Add the user to our request info
	attrs.User = types.User{
		Username: claims.Subject,
		Groups:   claims.Groups,
	}

	return nil
}

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
