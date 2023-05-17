package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/core/v2"
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

	// Determine the URL that serves this request so it can be later used as the
	// issuer URL
	ctx := context.WithValue(r.Context(), jwt.IssuerURLKey, issuerURL(r))

	client := api.NewAuthenticationClient(a.authenticator, a.store)
	tokens, err := client.CreateAccessToken(ctx, username, password)
	if err != nil {
		if err == corev2.ErrUnauthorized {
			logger.WithError(err).WithField("user", username).
				Error("invalid username and/or password")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		logger.WithError(err).Error("could not issue an access token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		logger.WithError(err).Error("couldn't write response")
	}
}

// test provides minimal username and password validation
func (a *AuthenticationRouter) test(w http.ResponseWriter, r *http.Request) {
	// Check for credentials provided in the Authorization header
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	client := api.NewAuthenticationClient(a.authenticator, a.store)
	err := client.TestCreds(r.Context(), username, password)
	if err == nil {
		return
	}

	logger.WithField(
		"user", username,
	).WithError(err).Info("invalid username and/or password")
	http.Error(w, "Request unauthorized", http.StatusUnauthorized)
}

// logout handles the logout flow
func (a *AuthenticationRouter) logout(w http.ResponseWriter, r *http.Request) {
	client := api.NewAuthenticationClient(a.authenticator, a.store)
	if err := client.Logout(r.Context()); err == nil {
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// token handles logic for issuing new access tokens
func (a *AuthenticationRouter) token(w http.ResponseWriter, r *http.Request) {
	client := api.NewAuthenticationClient(a.authenticator, a.store)

	// Determine the URL that serves this request so it can be later used as the
	// issuer URL
	ctx := context.WithValue(r.Context(), jwt.IssuerURLKey, issuerURL(r))

	tokens, err := client.RefreshAccessToken(ctx)
	if err != nil {
		if err == corev2.ErrInvalidToken {
			http.Error(w, "invalid access token", http.StatusBadRequest)
			return
		}

		logger.WithError(err).Info("unexpected error while authorizing refresh token")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		logger.WithError(err).Error("couldn't write response body")
	}
}

// issuerURL determines the URL used by the client to authenticate and renew its
// access token, so it can be later re-used by the client in case the access
// token is used against a different cluster
func issuerURL(r *http.Request) string {
	issuerURL := r.Host
	if r.TLS == nil {
		issuerURL = "http://" + issuerURL
	} else {
		issuerURL = "https://" + issuerURL
	}
	return issuerURL
}
