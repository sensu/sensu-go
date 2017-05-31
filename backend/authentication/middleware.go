package authentication

import (
	"net/http"
	"strings"
)

// Middleware is a HTTP middleware that enforces authentication
func Middleware(provider Provider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is enabled
		if !provider.AuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Does credentials were provided in the Authorization header?
		username, password, ok := r.BasicAuth()
		if ok {
			user, err := provider.Authenticate(username, password)
			if err != nil {
				logger.WithField(
					"user", username,
				).Info("Authentication failed. Invalid username or password")
				http.Error(w, "Request unauthorized", http.StatusUnauthorized)
				return
			}

			// Create the token and a signed version
			token, tokenString, err := newToken(user)
			if err != nil {
				logger.WithField(
					"user", username,
				).Infof("Authentication failed: %s", err.Error())
				http.Error(w, "Request unauthorized", http.StatusUnauthorized)
				return
			}

			// Set the claims into the request context
			setClaimsIntoContext(r, token)

			// Add the signed token to the response if login
			if r.URL.Path == "/auth" {
				w.Write([]byte(tokenString))
			}

			next.ServeHTTP(w, r)
			return
		}

		// Does a bearer token was provided in the Authorization header?
		var tokenString string
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			tokenString = tokens[0]
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		if tokenString != "" {
			token, err := parseToken(tokenString)
			if err != nil {
				logger.Infof("Authentication failed, error parsing token: %s", err.Error())
				http.Error(w, "Request unauthorized", http.StatusUnauthorized)
				return
			}

			// Set the claims into the request context
			setClaimsIntoContext(r, token)

			next.ServeHTTP(w, r)
			return
		}

		// The user is not authenticated
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	})
}
