package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

var (
	basicAuthRegex = regexp.MustCompile("^Basic ")
)

// RefreshToken middleware retrieves and validates a refresh token, provided
// in the body of a request, against an access token and the access list. Then,
// it adds the claims of both access and refresh tokens into the request
// context for easier consumption later
type RefreshToken struct{}

// Then ...
func (m RefreshToken) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the bearer token
		accessTokenString := jwt.ExtractBearerToken(r)

		// Ignore authorization header if BASIC auth
		if basicAuthRegex.MatchString(accessTokenString) {
			// http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			next.ServeHTTP(w, r)
			return
		}

		// We first need to validate that the access token is still valid, even if
		// it's expired
		accessToken, err := jwt.ValidateExpiredToken(accessTokenString)
		if err != nil {
			logger.WithError(err).Error("access token is invalid")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Retrieve the claims for the access token
		accessClaims, err := jwt.GetClaims(accessToken)
		if err != nil {
			logger.WithError(err).Error("could not parse the access token claims")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if accessClaims.Provider.ProviderType == "oidc" {
			http.Redirect(w, r, "/api/enterprise/authentication/v2/oidc/token", http.StatusTemporaryRedirect)
			return
		}

		decoder := json.NewDecoder(r.Body)
		payload := &v2.Tokens{}
		err = decoder.Decode(payload)
		if err != nil {
			logger.WithError(err).Error("could not decode the refresh token")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Now we want to validate the refresh token
		refreshToken, err := jwt.ValidateToken(payload.Refresh)
		if err != nil {
			logger.WithError(err).Error("refresh token is invalid")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		// Retrieve the claims for the refresh token
		refreshClaims, err := jwt.GetClaims(refreshToken)
		if err != nil {
			logger.WithError(err).Error("could not parse the refresh token claims")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Make sure the refresh token belongs to the same user as the access token
		if accessClaims.Subject == "" || accessClaims.Subject != refreshClaims.Subject {
			logger.WithFields(logrus.Fields{
				"user":          refreshClaims.Subject,
				"access_token":  accessClaims.Subject,
				"refresh_token": refreshClaims.Subject,
			}).Error("the access and refresh tokens subject do not match")
			http.Error(w, "Request unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), v2.AccessTokenClaims, accessClaims)
		ctx = context.WithValue(ctx, v2.RefreshTokenClaims, refreshClaims)
		ctx = context.WithValue(ctx, v2.RefreshTokenString, payload.Refresh)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
