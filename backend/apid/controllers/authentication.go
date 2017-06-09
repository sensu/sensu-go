package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/types"
)

// AuthenticationController handles authentication related requests
type AuthenticationController struct {
	Provider authentication.Provider
}

// Register the EventsController with a mux.Router.
func (a *AuthenticationController) Register(r *mux.Router) {
	r.HandleFunc("/auth", a.login).Methods(http.MethodGet)
	r.HandleFunc("/auth/token", a.token).Methods(http.MethodPost)
}

// login handles the login flow
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
		).Infof("Authentication failed, could not issue an access token: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve the claims because we later need the expiration
	claims, err := jwt.GetClaims(token)
	if err != nil {
		logger.WithField(
			"user", username,
		).Infof("Authentication failed, could not parse the access token claims: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	refreshTokenString, err := jwt.RefreshToken(user.Username)
	if err != nil {
		logger.WithField(
			"user", username,
		).Infof("Authentication failed, could not issue a refresh token: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(resBytes))
}

// token handles logic for issuing new access tokens
func (a *AuthenticationController) token(w http.ResponseWriter, r *http.Request) {
	// Retrieve the bearer token
	accessTokenString := jwt.ExtractBearerToken(r)
	if accessTokenString == "" {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// We first need to validate that the access token is still valid, even if
	// it's expired
	accessToken, err := jwt.ValidateExpiredToken(accessTokenString)
	if err != nil {
		logger.Infof("The access token is invalid: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve the refresh token
	if r.Body == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := &types.Tokens{}
	err = decoder.Decode(payload)
	if err != nil {
		logger.Infof("Could not decode the refresh token: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Now we want to validate the refresh token
	refreshToken, err := jwt.ValidateToken(payload.Refresh)
	if err != nil {
		logger.Infof("The refresh token is invalid: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve the claims for both tokens
	accessClaims, err := jwt.GetClaims(accessToken)
	if err != nil {
		logger.Infof("Could not parse the access token claims: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	refreshClaims, err := jwt.GetClaims(refreshToken)
	if err != nil {
		logger.Infof("Could not parse the refresh token claims: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure the refresh token belongs to the same user as the access token
	if accessClaims.Subject == "" || accessClaims.Subject != refreshClaims.Subject {
		logger.WithField(
			"user", refreshClaims.Subject,
		).Infof("The access and refresh tokens subject does not match: %s != %s",
			accessClaims.Subject,
			refreshClaims.Subject,
		)
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Issue a new access token
	accessToken, accessTokenString, err = jwt.AccessToken(refreshClaims.Subject)
	if err != nil {
		logger.WithField(
			"user", refreshClaims.Subject,
		).Infof("Could not issue a new access token: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve the claims because we later need the expiration
	accessClaims, err = jwt.GetClaims(accessToken)
	if err != nil {
		logger.WithField(
			"user", refreshClaims.Subject,
		).Infof("Could not issue a new access token: %s", err.Error())
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}

	// Prepare the response body
	response := &types.Tokens{
		Access:    accessTokenString,
		ExpiresAt: accessClaims.ExpiresAt,
		Refresh:   payload.Refresh,
	}

	resBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(resBytes))
}
