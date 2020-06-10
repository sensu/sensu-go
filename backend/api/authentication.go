package api

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
)

// AuthenticationClient is an API client for authentication.
type AuthenticationClient struct {
	auth *authentication.Authenticator
}

// NewAuthenticationClient creates a new AuthenticationClient, given a a store
// and an authenticator.
func NewAuthenticationClient(auth *authentication.Authenticator) *AuthenticationClient {
	return &AuthenticationClient{
		auth: auth,
	}
}

// CreateAccessToken creates a new access token, given a valid username and
// password.
func (a *AuthenticationClient) CreateAccessToken(ctx context.Context, username, password string) (*corev2.Tokens, error) {
	claims, err := a.auth.Authenticate(ctx, username, password)
	if err != nil {
		return nil, corev2.ErrUnauthorized
	}

	// Add the 'system:users' group to this user
	claims.Groups = append(claims.Groups, "system:users")

	// Add the issuer URL
	if issuer := ctx.Value(jwt.IssuerURLKey); issuer != nil {
		claims.Issuer = issuer.(string)
	}

	// Create an access token and its signed version
	_, tokenString, err := jwt.AccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("error creating access token: %s", err)
	}

	// Create a refresh token and its signed version
	refreshClaims := &corev2.Claims{StandardClaims: corev2.StandardClaims(claims.Subject)}
	_, refreshTokenString, err := jwt.RefreshToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("error creating access token: %s", err)
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
		} else {
			return err
		}
	}

	return errors.New("basic provider is disabled")
}

// Logout logs a user out. The context must carry the user's access and refresh
// claims, with the following context key-values:
//
// corev2.AccessTokenClaims -> *corev2.Claims
// corev2.RefreshTokenClaims -> *corev2.Claims
func (a *AuthenticationClient) Logout(ctx context.Context) error {
	return nil
}

// RefreshAccessToken refreshes an access token. The context must carry the
// user's access and refresh claims, as well as the previous token value,
// with the following context key-values:
//
// corev2.AccessTokenClaims -> *corev2.Claims
// corev2.RefreshTokenClaims -> *corev2.Claims
// corev2.RefreshTokenString -> string
func (a *AuthenticationClient) RefreshAccessToken(ctx context.Context) (*corev2.Tokens, error) {
	var accessClaims *corev2.Claims

	// Get the access token claims
	if value := ctx.Value(corev2.AccessTokenClaims); value != nil {
		accessClaims = value.(*corev2.Claims)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	// Get the refresh token claims
	if value := ctx.Value(corev2.RefreshTokenClaims); value == nil {
		return nil, corev2.ErrInvalidToken
	}

	// Get the refresh token string
	var refreshTokenString string
	if value := ctx.Value(corev2.RefreshTokenString); value != nil {
		refreshTokenString = value.(string)
	} else {
		return nil, corev2.ErrInvalidToken
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
	claims, err := a.auth.Refresh(ctx, accessClaims)
	if err != nil {
		return nil, err
	}

	// Ensure the 'system:users' group is present
	claims.Groups = append(claims.Groups, "system:users")

	// Add the issuer URL
	if issuer := ctx.Value(jwt.IssuerURLKey); issuer != nil {
		claims.Issuer = issuer.(string)
	}

	// Issue a new access token
	_, accessTokenString, err := jwt.AccessToken(claims)
	if err != nil {
		return nil, err
	}

	return &corev2.Tokens{
		Access:    accessTokenString,
		ExpiresAt: claims.ExpiresAt,
		Refresh:   refreshTokenString,
	}, nil
}
