package middlewares

import (
	"net/http"
	"strings"

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
// The general, expected format of a path is one of the following:
// /apis/{group}/{version}/namespaces, to create or list namespaces
// /apis/{group}/{version}/namespaces/{name}
// /apis/{group}/{version}/namespaces/{namespace}/{kind}
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

// LegacyAuthorizationAttributes is an HTTP middleware that populates a context
// for the request, to be used by other middlewares. You probably want this
// middleware to be executed early in the middleware stack.
//
// This middleware is here purely to support, old, pre-API versioning routes. It
// should be removed when all the old routes are replaced by versioned ones.
//
// The expected path is one of the following:
// /assets
// /assets/{id}
// /checks
// /checks/{id}
// /checks/{id}/execute
// /checks/{id}/hooks/{type}
// /checks/{id}/hooks/{type}/hook/{hook}
// /cluster/members
// /cluster/members/{id}
// /entities
// /entities/{id}
// /filters
// /filters/{id}
// /events
// /events/{entity}
// /events/{entity}/{check}
// /extensions
// /extensions/{id}
// /handlers
// /handlers/{id}
// /hooks
// /hooks/{id}
// /mutators
// /mutators/{id}
// /rbac/namespaces
// /rbac/namespaces/{id}
// /rbac/users
// /rbac/users/{id}
// /silenced
// /silenced{id}
// /silenced/subscriptions
// /silenced/subscriptions/{subscription}
// /silenced/checks
// /silenced/checks/{check}
type LegacyAuthorizationAttributes struct{}

func (a LegacyAuthorizationAttributes) Then(next http.Handler) http.Handler {
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
		attrs.APIGroup = "core"
		attrs.APIVersion = "v2"

		// In the legacy routes, a non-default namespace is passed as a URL
		// query parameter like so: ?namespace=foo
		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "default"
		}
		attrs.Namespace = namespace

		// The resource type is always the first element of the path, except for
		// namespaces (/rbac/namespaces), users (/rbac/users) and cluster
		// members (/cluster/members).
		pathParts := strings.Split(r.URL.Path, "/")

		switch pathParts[0] {
		case "cluster":
			attrs.Resource = "cluster-members"
		case "rbac":
			if len(pathParts) >= 2 {
				attrs.Resource = pathParts[1]
			}
		default:
			attrs.Resource = pathParts[0]
		}

		// In the legacy routes, the resource name is always a route variable
		// named "id", except for entities.
		if attrs.Resource == "entities" {
			attrs.ResourceName = vars["entity"]
		} else {
			attrs.ResourceName = vars["id"]
		}

		if attrs.Verb == "get" && attrs.ResourceName == "" {
			attrs.Verb = "list"
		}
	})
}
