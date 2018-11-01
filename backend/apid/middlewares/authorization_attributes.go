package middlewares

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
)

// AuthorizationAttributes is an HTTP middleware that populates a context for the
// request, to be used by other middlewares. You probably want this middleware
// to be executed early in the middleware stack.
type AuthorizationAttributes struct{}

// Then is the AuthorizationAttrs middleware's main logic, heavily inspired by
// Kubernetes' WithRequestInfo.
//
// The general, expected format of a path one of the following: -
// /apis/{group}/{version}/namespaces, to create or list namespaces -
// /apis/{group}/{version}/namespaces/{name} -
// /apis/{group}/{version}/namespaces/{namespace}/{kind} -
// /apis/{group}/{version}/namespaces/{namespace}/{kind}/{name}
//
// This middleware tries to fill in as many fields as it can in the
// authorization.Attributes struct added to the context, but there is no
// guarantee of any fields being filled in.
func (a AuthorizationAttributes) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		attrs := &authorization.Attributes{}

		ctx = authorization.SetAttributes(ctx, attrs)
		defer next.ServeHTTP(w, r.WithContext(ctx))

		switch r.Method {
		case "POST":
			attrs.Verb = "create"
		case "GET", "HEAD":
			attrs.Verb = "get"
		case "PUT":
			attrs.Verb = "update"
		case "DELETE":
			attrs.Verb = "delete"
		default:
			attrs.Verb = ""
		}

		vars := mux.Vars(r)
		attrs.APIGroup = vars["group"]
		attrs.APIVersion = vars["version"]
		attrs.Namespace = vars["namespace"]
		attrs.Resource = vars["kind"]
		attrs.ResourceName = vars["name"]

		if attrs.Verb == "get" && attrs.ResourceName == "" {
			attrs.Verb = "list"
		}
	})
}
