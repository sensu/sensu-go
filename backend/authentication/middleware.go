package authentication

import (
	"net/http"
)

// Middleware is a HTTP middleware that enforces authentication
func Middleware(provider Provider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is enabled
		if !provider.AuthEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// Does credentials were provided in the request's Authorization header?
		username, password, ok := r.BasicAuth()
		if ok {
			_, err := provider.Authenticate(username, password)
			if err != nil {
				http.Error(w, "Request unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		// The user is not authenticated
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	})
}
