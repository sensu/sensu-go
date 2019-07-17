package routers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

	client := api.NewAuthenticationClient(a.store, a.authenticator)
	tokens, err := client.CreateAccessToken(r.Context(), username, password)

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

	client := api.NewAuthenticationClient(a.store, a.authenticator)
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
	client := api.NewAuthenticationClient(a.store, a.authenticator)
	err := client.Logout(r.Context())
	if err == nil {
		return
	}

	if err == corev2.ErrInvalidToken {
		http.Error(w, "invalid refresh token", http.StatusBadRequest)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// token handles logic for issuing new access tokens
func (a *AuthenticationRouter) token(w http.ResponseWriter, r *http.Request) {
	client := api.NewAuthenticationClient(a.store, a.authenticator)
	tokens, err := client.RefreshAccessToken(r.Context())
	if err != nil {
		if err == corev2.ErrInvalidToken {
			http.Error(w, "invalid access token", http.StatusBadRequest)
			return
		}
		if _, ok := err.(*store.ErrNotFound); ok {
			logger.WithError(err).Info("refresh token unauthorized")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		logger.WithError(err).Info("unexpected error while authorizing refresh token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		logger.WithError(err).Error("couldn't write response body")
	}
}
