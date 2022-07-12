package agentd

import (
	"net/http"

	"github.com/gorilla/mux"
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
