package apid

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
)

// WrappedRouter is equivelant to mux.Router with the addition that any time it
// handles a request the given middleware stack is applied first.
type WrappedRouter struct {
	*mux.Router
	middleware *middlewares.Stack
}

// Match matches registered routes against the request.
func (r *WrappedRouter) Match(req *http.Request, match *mux.RouteMatch) bool {
	if !r.Router.Match(req, match) {
		return false
	}

	// wrap the handler that matched the request in our middleware stack
	match.Handler = middlewares.Apply(match.Handler, r.middleware)

	return true
}

// NewSubrouter returns new router given route and set of middleware. Useful
// when you would like a subset of routes that are wrapped in a specific set of
// middleware. Eg. you would some routes to require authetication while others
// you would not.
func NewSubrouter(r *mux.Route, ms ...middlewares.HTTPMiddleware) *mux.Router {
	subRouter := r.Subrouter().UseEncodedPath()
	wrapper := WrappedRouter{
		Router:     subRouter,
		middleware: middlewares.NewStack(ms),
	}

	// Override existing matcher for the subrouter w/ variant wrapped in
	// middleware stack.
	r.MatcherFunc(wrapper.Match)

	return subRouter
}
