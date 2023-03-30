package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// Authentication is a HTTP middleware that enforces authentication
type Authentication struct {
	// IgnoreUnauthorized configures the middleware to continue the handler chain
	// in the case where an access token was not present.
	IgnoreUnauthorized bool
	Store              storev2.Interface
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

func extractAPIKeyClaims(ctx context.Context, key string, store storev2.Interface) (*corev2.Claims, error) {
	var claims *corev2.Claims
	keyStore := storev2.Of[*corev2.APIKey](store)
	apiKeys, err := keyStore.List(ctx, storev2.ID{}, nil)
	if err != nil {
		return nil, err
	}

	for _, apiKey := range apiKeys {
		if bcrypt.CheckPassword(string(apiKey.Hash), key) {
			userStore := storev2.Of[*corev2.User](store)
			user, err := userStore.Get(ctx, storev2.ID{Name: apiKey.Username})
			if err != nil {
				return nil, err
			}

			// inject the username and groups into standard jwt claims
			claims = &corev2.Claims{
				StandardClaims: corev2.StandardClaims(user.Username),
				Groups:         user.Groups,
				APIKey:         true,
			}

			return claims, nil
		}
	}

	return nil, errors.New("API key rejected")
}

type errorWriter struct {
	err actions.Error
}

// Then middleware
func (e errorWriter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, e.err)
	})
}
