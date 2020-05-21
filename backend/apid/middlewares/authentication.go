package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct {
	// IgnoreUnauthorized configures the middleware to continue the handler chain
	// in the case where an access token was not present.
	IgnoreUnauthorized bool
	Store              store.Store
}

// Then middleware
func (a Authentication) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader, ok := r.Header["Authorization"]
		if ok && len(authHeader) >= 1 {
			headerString := authHeader[0]

			// if the auth header contains Bearer, continue with token auth
			if strings.HasPrefix(headerString, "Bearer ") {
				headerString = strings.TrimPrefix(headerString, "Bearer ")
				token, err := jwt.ValidateToken(headerString)
				if err != nil {
					logger.WithError(err).Warn("invalid token")
					actionErr := actions.NewErrorf(actions.Unauthenticated, "invalid credentials")
					SimpleLogger{}.Then(errorWriter{err: actionErr}.Then(next)).ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// Set the claims into the request context
				ctx = jwt.SetClaimsIntoContext(r, token.Claims.(*corev2.Claims))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// if the auth header contains Key, continue with api key auth
			if strings.HasPrefix(headerString, "Key ") {
				headerString = strings.TrimPrefix(headerString, "Key ")
				claims, err := extractAPIKeyClaims(ctx, headerString, a.Store)
				if err != nil {
					logger.WithError(err).Warn("invalid api key")
					actionErr := actions.NewErrorf(actions.Unauthenticated, "invalid credentials")
					SimpleLogger{}.Then(errorWriter{err: actionErr}.Then(next)).ServeHTTP(w, r.WithContext(ctx))
					return
				}
				if claims != nil {
					// Set the claims into the request context
					ctx = jwt.SetClaimsIntoContext(r, claims)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// The user is not authenticated
		if a.IgnoreUnauthorized {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		actionErr := actions.NewErrorf(actions.Unauthenticated, "bad credentials")
		SimpleLogger{}.Then(errorWriter{err: actionErr}.Then(next)).ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractAPIKeyClaims(ctx context.Context, key string, store store.Store) (*corev2.Claims, error) {
	var claims *corev2.Claims
	// retrieve the APIKey based on the key provided
	apiKey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: key,
		},
	}
	if err := store.GetResource(context.Background(), apiKey.Name, apiKey); err != nil {
		return claims, err
	}

	// retrieve the sensu user associated with the key provided
	user, err := store.GetUser(ctx, apiKey.Username)
	if err != nil {
		return claims, err
	}
	// this shouldn't happen because user validation happens at key generation,
	// but in the event a user is deleted after a key has been generated,
	// the key should not pass authentication
	if user == nil {
		return claims, fmt.Errorf("user %s not found", apiKey.Username)
	}

	// inject the username and groups into standard jwt claims
	claims = &corev2.Claims{
		StandardClaims: corev2.StandardClaims(user.Username),
		Groups:         user.Groups,
		APIKey:         true,
	}

	return claims, nil
}

type errorWriter struct {
	err actions.Error
}

// Then middleware
func (e errorWriter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, e.err)
		return
	})
}
