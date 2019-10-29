package agentd

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// AuthenticationMiddleware represents the middleware used for authentication
var AuthenticationMiddleware mux.MiddlewareFunc

// AuthorizationMiddleware represents the middleware used for authorization
var AuthorizationMiddleware mux.MiddlewareFunc

// AgentLimiterMiddleware represents the middleware used to limit agent
// sessions if the entity limit has been reached.
var AgentLimiterMiddleware mux.MiddlewareFunc

// EntityLimiterMiddleware represents the middleware used to add entity limit
// headers if the entity limit has been reached.
var EntityLimiterMiddleware mux.MiddlewareFunc

// authenticate is the abstraction layer required to be able to change at
// runtime the actual function assigned to AuthenticationMiddleware above
func authenticate(next http.Handler) http.Handler {
	return AuthenticationMiddleware(next)
}

// authorize is the abstraction layer required to be able to change at
// runtime the actual function assigned to AuthenticationMiddleware above
func authorize(next http.Handler) http.Handler {
	return AuthorizationMiddleware(next)
}

// agentLimit is the abstraction layer required to be able to change at
// runtime the actual function assigned to AgentLimiterMiddleware above.
func agentLimit(next http.Handler) http.Handler {
	return AgentLimiterMiddleware(next)
}

// entityLimit is the abstraction layer required to be able to change at
// runtime the actual function assigned to EntityLimiterMiddleware above.
func entityLimit(next http.Handler) http.Handler {
	return EntityLimiterMiddleware(next)
}

// AuthStore specifies the storage requirements for authentication and
// authorization.
type AuthStore interface {
	// AuthenticateUser attempts to authenticate a user with the given username
	// and hashed password. An error is returned if the user does not exist, is
	// disabled or the given password does not match.
	AuthenticateUser(ctx context.Context, user, pass string) (*corev2.User, error)

	// GetUser directly retrieves a user with the given username. An error is
	// returned if the user does not exist or is disabled
	GetUser(ctx context.Context, username string) (*corev2.User, error)
}
