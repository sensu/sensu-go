package routers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sensu/sensu-go/api/core/v2"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// AuthenticationRouter handles authentication related requests
type AuthenticationRouter struct {
	store         store.Store
	authenticator *authentication.Authenticator
}

// NewAuthenticationRouter instantiates new router.
func NewAuthenticationRouter(store store.Store, authenticator *authentication.Authenticator) *AuthenticationRouter {
	return &AuthenticationRouter{store: store, authenticator: authenticator}
}

// Mount the authentication routes on given mux.Router.
func (a *AuthenticationRouter) Mount(r *mux.Router) {
	r.HandleFunc("/auth", a.login).Methods(http.MethodGet)
	r.HandleFunc("/auth/test", a.test).Methods(http.MethodGet)
	r.HandleFunc("/auth/token", a.token).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", a.logout).Methods(http.MethodPost)
}

// login handles the login flow
func (a *AuthenticationRouter) login(w http.ResponseWriter, r *http.Request) {
	// Check for credentials provided in the Authorization header
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Authenticate against the provider
	claims, err := a.authenticator.Authenticate(r.Context(), username, password)
	if err != nil {
		logger.WithError(err).WithField("user", username).
			Error("invalid username and/or password")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	logger = logger.WithField("user", claims.Subject)

	// Add the 'system:users' group to this user
	claims.Groups = append(claims.Groups, "system:users")

	// Create an access token and its signed version
	token, tokenString, err := jwt.AccessToken(claims)
	if err != nil {
		logger.WithError(err).Error("could not issue an access token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Create a refresh token and its signed version
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	refreshToken, refreshTokenString, err := jwt.RefreshToken(refreshClaims)
	if err != nil {
		logger.WithError(err).Error("could not issue a refresh token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Add the tokens in the access list
	if err := a.store.AllowTokens(token, refreshToken); err != nil {
		logger.WithError(err).Error("could not add tokens to the access list")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Prepare the response body
	response := &types.Tokens{
		Access:    tokenString,
		ExpiresAt: claims.ExpiresAt,
		Refresh:   refreshTokenString,
	}

	resBytes, err := json.Marshal(response)
	if err != nil {
		logger.WithError(err).Error("could not encode response body")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, string(resBytes))
}

// test provides minimal username and password validation
func (a *AuthenticationRouter) test(w http.ResponseWriter, r *http.Request) {
	// Check for credentials provided in the Authorization header
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Authenticate against the basic provider
	var err error
	providers := a.authenticator.Providers()

	if basic, ok := providers[basic.Type]; ok {
		_, err = basic.Authenticate(r.Context(), username, password)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		err = errors.New("basic provider is disabled")
	}

	logger.WithField(
		"user", username,
	).WithError(err).Info("invalid username and/or password")
	http.Error(w, "Request unauthorized", http.StatusUnauthorized)
}

// logout handles the logout flow
func (a *AuthenticationRouter) logout(w http.ResponseWriter, r *http.Request) {
	var accessClaims, refreshClaims *types.Claims

	// Get the access token claims
	if value := r.Context().Value(types.AccessTokenClaims); value != nil {
		accessClaims = value.(*v2.Claims)
	} else {
		http.Error(w, "invalid access token", http.StatusBadRequest)
		return
	}

	// Get the refresh token claims
	if value := r.Context().Value(types.RefreshTokenClaims); value != nil {
		refreshClaims = value.(*v2.Claims)
	} else {
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
		return
	}

	// Remove the access & refresh tokens from the access list
	if err := a.store.RevokeTokens(accessClaims, refreshClaims); err != nil {
		logger.WithError(err).WithField("user", refreshClaims.Subject).Errorf(
			"could not revoke tokens IDs %q & %q", accessClaims.Id, refreshClaims.Id,
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// token handles logic for issuing new access tokens
func (a *AuthenticationRouter) token(w http.ResponseWriter, r *http.Request) {
	var accessClaims, refreshClaims *v2.Claims

	// Get the access token claims
	if value := r.Context().Value(v2.AccessTokenClaims); value != nil {
		accessClaims = value.(*v2.Claims)
	} else {
		http.Error(w, "invalid access token", http.StatusBadRequest)
		return
	}

	logger = logger.WithField("user", accessClaims.Subject)

	// Get the refresh token claims
	if value := r.Context().Value(v2.RefreshTokenClaims); value != nil {
		refreshClaims = value.(*v2.Claims)
	} else {
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
		return
	}

	// Get the refresh token string
	var refreshTokenString string
	if value := r.Context().Value(v2.RefreshTokenString); value != nil {
		refreshTokenString = value.(string)
	} else {
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
		return
	}

	// Make sure the refresh token is authorized in the access list
	if _, err := a.store.GetToken(refreshClaims.Subject, refreshClaims.Id); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			logger.WithError(err).Infof("refresh token ID %q is unauthorized", refreshClaims.Id)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		default:
			logger.WithError(err).Infof("unexpected error while authorizing refresh token ID %q", refreshClaims.Id)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// Revoke the old access token from the access list
	if err := a.store.RevokeTokens(accessClaims); err != nil {
		logger.WithError(err).Errorf("could not revoke access token ID %q", accessClaims.Id)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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
	claims, err := a.authenticator.Refresh(r.Context(), accessClaims.Provider)
	if err != nil {
		logger.WithError(err).Error("could not refresh user claims")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Ensure the 'system:users' group is present
	claims.Groups = append(claims.Groups, "system:users")

	// Issue a new access token
	accessToken, accessTokenString, err := jwt.AccessToken(claims)
	if err != nil {
		logger.WithError(err).Error("could not issue a new access token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// store the new access token in the access list
	if err = a.store.AllowTokens(accessToken); err != nil {
		logger.WithError(err).Errorf("could not allow new access token %q", claims.Id)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Prepare the response body
	response := &types.Tokens{
		Access:    accessTokenString,
		ExpiresAt: accessClaims.ExpiresAt,
		Refresh:   refreshTokenString,
	}

	resBytes, err := json.Marshal(response)
	if err != nil {
		logger.WithError(err).Error("could not encode response body")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, string(resBytes))
}
