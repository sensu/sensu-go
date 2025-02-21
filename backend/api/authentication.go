package api

import (
	"context"
	"errors"
	"fmt"
	corev2 "github.com/sensu/core/v2"
	"time"

	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"
)

// AuthenticationClient is an API client for authentication.
type AuthenticationClient struct {
	auth         *authentication.Authenticator
	sessionStore store.SessionStore
}

// NewAuthenticationClient creates a new AuthenticationClient, given an
// authenticator and a session store.
func NewAuthenticationClient(auth *authentication.Authenticator, sessionStore store.SessionStore) *AuthenticationClient {
	return &AuthenticationClient{
		auth:         auth,
		sessionStore: sessionStore,
	}
}

// CreateAccessToken creates a new access token, given a valid username and
// password.
func (a *AuthenticationClient) CreateAccessToken(ctx context.Context, username, password string) (*corev2.Tokens, error) {
	claims, err := a.auth.Authenticate(ctx, username, password)
	if err != nil {
		return nil, corev2.ErrUnauthorized
	}

	// Initialize a new session for this user
	sessionID, err := jwt.InitSession(claims.Subject)
	if err != nil {
		return nil, err
	}
	claims.SessionID = sessionID

	// Add the 'system:users' group to this user
	claims.Groups = append(claims.Groups, "system:users")

	// Add the issuer URL
	if issuer := ctx.Value(jwt.IssuerURLKey); issuer != nil {
		claims.Issuer = issuer.(string)
	}

	// append configured access token expiry to claims
	var accessTokenExpiry time.Duration
	if accessTokenExp := ctx.Value("accessTokenExpiry"); accessTokenExp != nil {
		accessTokenExpiry = accessTokenExp.(time.Duration)
	}

	// Create an access token and its signed version
	_, tokenString, err := jwt.AccessToken(claims, jwt.WithAccessTokenExpiry(accessTokenExpiry))
	if err != nil {
		return nil, fmt.Errorf("error creating access token: %s", err)
	}

	// Create a refresh token and its signed version
	refreshClaims := &corev2.Claims{
		StandardClaims: corev2.StandardClaims(claims.Subject),
		SessionID:      sessionID,
	}

	// append configured refresh token expiry to claims
	var refreshTokenExpiry time.Duration
	if refreshTokenExp := ctx.Value("refreshTokenExpiry"); refreshTokenExp != nil {
		refreshTokenExpiry = refreshTokenExp.(time.Duration)
	}

	refreshToken, refreshTokenString, err := jwt.RefreshToken(refreshClaims, jwt.WithRefreshTokenExpiry(refreshTokenExpiry))
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token: %s", err)
	}

	refreshTokenClaims, err := jwt.GetClaims(refreshToken)
	if err != nil {
		return nil, err
	}

	// Store the refresh token's unique ID as part of this user's session
	if err := a.sessionStore.UpdateSession(ctx, refreshTokenClaims.Subject, refreshTokenClaims.SessionID, refreshTokenClaims.Id); err != nil {
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
//
// Given that we use JWTs for authentication, logging out just destroys the
// server side session such that it's not possible to get a new access token
// anymore, regardless of the refresh token presented by the user.
//
// Again, because we use JWTs, even after logging out, the access token bearer
// can still interact with the system until the token expires (up to 5 minutes
// by default).
func (a *AuthenticationClient) Logout(ctx context.Context) error {
	var accessClaims *corev2.Claims

	// Retrieve the access token's claims
	if value := ctx.Value(corev2.AccessTokenClaims); value != nil {
		accessClaims = value.(*corev2.Claims)
	} else {
		return corev2.ErrInvalidToken
	}

	return a.sessionStore.DeleteSession(ctx, accessClaims.Subject, accessClaims.SessionID)
}

// RefreshAccessToken refreshes an access/refresh token pair. The context must
// carry the user's access and refresh claims, as well as the previous token
// value, with the following context key-values:
//
// corev2.AccessTokenClaims -> *corev2.Claims
// corev2.RefreshTokenClaims -> *corev2.Claims
// corev2.RefreshTokenString -> string
func (a *AuthenticationClient) RefreshAccessToken(ctx context.Context) (*corev2.Tokens, error) {
	var accessClaims *corev2.Claims
	var refreshClaims *corev2.Claims

	// Get the access token claims
	if value := ctx.Value(corev2.AccessTokenClaims); value != nil {
		accessClaims = value.(*corev2.Claims)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	// Get the refresh token claims
	if value := ctx.Value(corev2.RefreshTokenClaims); value != nil {
		refreshClaims = value.(*corev2.Claims)
	} else {
		return nil, corev2.ErrInvalidToken
	}

	sessionID := accessClaims.SessionID

	storedRefreshTokenID, err := a.sessionStore.GetSession(ctx, refreshClaims.Subject, refreshClaims.SessionID)
	if err != nil {
		return nil, err
	}

	// If the supplied refresh token's ID doesn't match what the session
	// expected it to be. Whatever the reason for that, be it refresh token
	// reuse or otherwise, we just tear down that session, forcing the user to
	// fully reauthenticate.
	if refreshClaims.Id != storedRefreshTokenID {
		a.Logout(ctx)
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

	// Carry over the session ID
	claims.SessionID = sessionID

	// Ensure the 'system:users' group is present
	claims.Groups = append(claims.Groups, "system:users")

	// Add the issuer URL
	if issuer := ctx.Value(jwt.IssuerURLKey); issuer != nil {
		claims.Issuer = issuer.(string)
	}

	// append configured access token expiry to claims
	var accessTokenExpiry time.Duration
	if accessTokenExp := ctx.Value("accessTokenExpiry"); accessTokenExp != nil {
		accessTokenExpiry = accessTokenExp.(time.Duration)
	}

	// Issue a new access token
	_, newAccessTokenString, err := jwt.AccessToken(claims, jwt.WithAccessTokenExpiry(accessTokenExpiry))
	if err != nil {
		return nil, err
	}

	// append configured refresh token expiry to claims
	var refreshTokenExpiry time.Duration
	if refreshTokenExp := ctx.Value("refreshTokenExpiry"); refreshTokenExp != nil {
		refreshTokenExpiry = refreshTokenExp.(time.Duration)
	}

	// Create a new refresh token, carrying over the session ID
	newRefreshClaims := &corev2.Claims{
		StandardClaims: corev2.StandardClaims(claims.Subject),
		SessionID:      sessionID,
	}
	newRefreshToken, newRefreshTokenString, err := jwt.RefreshToken(newRefreshClaims, jwt.WithRefreshTokenExpiry(refreshTokenExpiry))
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token: %s", err)
	}

	newRefreshTokenClaims, err := jwt.GetClaims(newRefreshToken)
	if err != nil {
		return nil, err
	}

	// Update the session with the new refresh token's unique ID
	if err := a.sessionStore.UpdateSession(ctx, claims.Subject, refreshClaims.SessionID, newRefreshTokenClaims.Id); err != nil {
		return nil, err
	}

	return &corev2.Tokens{
		Access:    newAccessTokenString,
		ExpiresAt: claims.ExpiresAt,
		Refresh:   newRefreshTokenString,
	}, nil
}
