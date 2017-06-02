package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// authenticationSuccessResponse contains the structure for the response body
// of a successful authentication
type authenticationSuccessResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

// AuthenticationController handles authentication related requests
type AuthenticationController struct {
	Provider authentication.Provider
}

// Register the EventsController with a mux.Router.
func (a *AuthenticationController) Register(r *mux.Router) {
	r.HandleFunc("/auth", a.login).Methods(http.MethodGet)
}

func (a *AuthenticationController) login(w http.ResponseWriter, r *http.Request) {
	// Check for credentials provided in the Authorization header
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Authenticate against the provider
	user, err := a.Provider.Authenticate(username, password)
	if err != nil {
		logger.WithField(
			"user", username,
		).Info("Authentication failed. Invalid username or password")
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Create the token and a signed version
	token, tokenString, err := jwt.AccessToken(user.Username)
	if err != nil {
		logger.WithField(
			"user", username,
		).Infof("Authentication failed: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := jwt.GetClaims(token)
	if err != nil {
		logger.WithField(
			"user", username,
		).Infof("Authentication failed: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	response := &authenticationSuccessResponse{
		AccessToken: tokenString,
		ExpiresAt:   claims.ExpiresAt,
	}

	resBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(resBytes))
}
