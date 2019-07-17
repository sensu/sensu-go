package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// AuthenticationClient is a type that can be used to interact with the
// backend's authentication services.
type AuthenticationClient struct {
	store store.TokenStore
	auth  *authentication.Authenticator
}

func NewAuthenticationClient(store store.TokenStore, auth *authentication.Authenticator) *AuthenticationClient {
	return &AuthenticationClient{
		store: store,
		auth:  auth,
	}
}

// CreateAccessToken creates a new access token, given a valid username and password.
func (a *AuthenticationClient) CreateAccessToken(ctx context.Context, username, password string) (*corev2.Tokens, error) {
	claims, err := a.auth.Authenticate(ctx, username, password)
	if err != nil {
		return nil, corev2.ErrUnauthorized
	}

	// Add the 'system:users' group to this user
	claims.Groups = append(claims.Groups, "system:users")

	// Create an access token and its signed version
	token, tokenString, err := jwt.AccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("error creating access token: %s", err)
	}

	// Create a refresh token and its signed version
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	refreshToken, refreshTokenString, err := jwt.RefreshToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("error creating access token: %s", err)
	}

	// Add the tokens in the access list
	if err := a.store.AllowTokens(token, refreshToken); err != nil {
		return nil, err
	}

	result := &corev2.Tokens{
		Access:    tokenString,
		ExpiresAt: claims.ExpiresAt,
		Refresh:   refreshTokenString,
	}

	return result, nil

}

// TestCreds detects if the username and password are valid.
func (a *AuthenticationClient) TestCreds(ctx context.Context, username, password string) error {
	// Authenticate against the basic provider
	providers := a.auth.Providers()

	if basic, ok := providers[basic.Type]; ok {
		if _, err := basic.Authenticate(ctx, username, password); err == nil {
			return nil
		}
		return errors.New("basic provider is disabled")
	}

	return errors.New("invalid username and/or password")
}

// Logout logs a user out. The context must carry the user's access and refresh
// claims, with the following context key-values:
//
// corev2.AccessTokenClaims -> *corev2.Claims
// corev2.RefreshTokenClaims -> *corev2.Claims
func (a *AuthenticationClient) Logout(ctx context.Context) error {
	var accessClaims, refreshClaims *corev2.Claims

	// Get the access token claims
	if value := ctx.Value(corev2.AccessTokenClaims); value != nil {
		accessClaims = value.(*v2.Claims)
	} else {
		return corev2.ErrInvalidToken
	}

	// Get the refresh token claims
	if value := ctx.Value(corev2.RefreshTokenClaims); value != nil {
		refreshClaims = value.(*v2.Claims)
	} else {
		return corev2.ErrInvalidToken
	}

	// Remove the access & refresh tokens from the access list
	return a.store.RevokeTokens(accessClaims, refreshClaims)
}

// RefreshAccessToken refreshes an access token. The context must carry the
// user's access and refresh claims, as well as the previous token value,
// with the following context key-values:
//
// corev2.AccessTokenClaims -> *corev2.Claims
// corev2.RefreshTokenClaims -> *corev2.Claims
// corev2.RefreshTokenString -> string
func (a *AuthenticationClient) RefreshAccessToken(ctx context.Context) (*corev2.Tokens, error) {
	var accessClaims, refreshClaims *v2.Claims

	// Get the access token claims
	if value := ctx.Value(v2.AccessTokenClaims); value != nil {
		accessClaims = value.(*v2.Claims)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	// Get the refresh token claims
	if value := ctx.Value(v2.RefreshTokenClaims); value != nil {
		refreshClaims = value.(*v2.Claims)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	// Get the refresh token string
	var refreshTokenString string
	if value := ctx.Value(v2.RefreshTokenString); value != nil {
		refreshTokenString = value.(string)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	// Make sure the refresh token is authorized in the access list
	if _, err := a.store.GetToken(refreshClaims.Subject, refreshClaims.Id); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, err
		default:
			return nil, fmt.Errorf("unexpected error while authorizing refresh token ID %q: %s", refreshClaims.Id, err)
		}
	}

	// Revoke the old access token from the access list
	if err := a.store.RevokeTokens(accessClaims); err != nil {
		return nil, err
	}

	// Ensure backward compatibility by filling the provider claims if missing
	if accessClaims.Provider.ProviderID == "" || accessClaims.Provider.UserID == "" {
		accessClaims.Provider.ProviderID = basic.Type
		accessClaims.Provider.UserID = accessClaims.Subject
	}

	// Ensure backward compatibility with Sensu Go 5.1.1, since the provider type
	// was used in the provider ID
	if accessClaims.Provider.ProviderID == "basic/default" {
		accessClaims.Provider.ProviderID = basic.Type
	}

	// Refresh the user claims
	claims, err := a.auth.Refresh(ctx, accessClaims.Provider)
	if err != nil {
		return nil, corev2.ErrUnauthorized
	}

	// Ensure the 'system:users' group is present
	claims.Groups = append(claims.Groups, "system:users")

	// Issue a new access token
	accessToken, accessTokenString, err := jwt.AccessToken(claims)
	if err != nil {
		return nil, err
	}

	// store the new access token in the access list
	if err := a.store.AllowTokens(accessToken); err != nil {
		return nil, err
	}

	return &corev2.Tokens{
		Access:    accessTokenString,
		ExpiresAt: accessClaims.ExpiresAt,
		Refresh:   refreshTokenString,
	}, nil
}
