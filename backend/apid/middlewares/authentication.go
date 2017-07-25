package middlewares

import (
	"fmt"
	"net/http"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
)

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct{}

// Then middleware
func (a Authentication) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r.Header)
		tokenString := jwt.ExtractBearerToken(r)
		if tokenString != "" {
			token, err := jwt.ValidateToken(tokenString)
			fmt.Printf("Token is: %s, Error is %s", token, err)
			if err == nil {
				// Set the claims into the request context
				ctx := jwt.SetClaimsIntoContext(r.Context(), token)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// The user is not authenticated
		http.Error(w, "Bad credentials given", http.StatusUnauthorized)
		return
	})
}
