package middlewares

import (
	"context"
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
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
// /apis/{group}/{version}/namespaces/{namespace}/{type}
// /apis/{group}/{version}/namespaces/{namespace}/{type}/{name}
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
		attrs.Resource = vars["type"]
		attrs.ResourceName = vars["name"]

		if attrs.Verb == "get" && attrs.ResourceName == "" {
			attrs.Verb = "list"
		}

		// Add the user to the attributes
		if err := getUser(ctx, attrs); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
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

		// A non-default namespace is passed as a query parameter called "namespace"
		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "default"
		}
		attrs.Namespace = namespace

		// Add the user to the attributes
		if err := getUser(ctx, attrs); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// The resource type is always the first element of the path, except for
		// namespaces, users and cluster members.
		fullPath := r.URL.Path
		pathParts := strings.Split(strings.Trim(fullPath, "/"), "/")
		attrs.Resource = pathParts[0]

		switch attrs.Resource {
		case "cluster":
			attrs.Resource = "cluster-members"
		case "rbac":
			if len(pathParts) >= 2 {
				attrs.Resource = pathParts[1]
			}
		}

		// Most resource names are identified by a route variable named "id".
		// Other resources have snowflake paths; see their corresponding router
		// and the expected paths above.
		attrs.ResourceName = vars["id"]

		switch attrs.Resource {
		case "events":
			attrs.ResourceName = path.Join(vars["entity"], vars["check"])
		case "silenced":
			if strings.HasPrefix(fullPath, "/silenced/checks") {
				attrs.ResourceName = path.Join("checks", vars["check"])
			} else if strings.HasPrefix(fullPath, "/silenced/subscriptions") {
				attrs.ResourceName = path.Join("subscriptions", vars["subscription"])
			}
		}

		if attrs.Verb == "get" && (attrs.ResourceName == "" || isListable(attrs.Resource, attrs.ResourceName)) {
			attrs.Verb = "list"
		}

		// Verify if the authenticated user is trying to access itself
		if attrs.Resource == "users" && attrs.ResourceName == attrs.User.Username {
			// Change the resource to LocalSelfUserResource if a user views itself
			if attrs.Verb == "get" && vars["subresource"] == "" {
				attrs.Resource = types.LocalSelfUserResource
			}

			// Change the resource to LocalSelfUserResource if a user tries to change
			// its own password
			if attrs.Verb == "update" && vars["subresource"] == "password" {
				attrs.Resource = types.LocalSelfUserResource
			}
		}
	})
}

func isListable(resourceType, name string) bool {
	// For /events, if the resource name doesn't contain a '/', we're listing
	// /events/{entity} as opposed to getting /events/{entity}/{check}
	if resourceType == "events" && !strings.ContainsRune(name, '/') {
		return true
	}

	if resourceType == "silenced" && (name == "checks" || name == "subscriptions") {
		return true
	}

	return false
}

func getUser(ctx context.Context, attrs *authorization.Attributes) error {
	// Get the claims from the request context
	claims := jwt.GetClaimsFromContext(ctx)
	if claims == nil {
		return errors.New("no claims found in the request context")
	}

	// Add the user to our request info
	attrs.User = types.User{
		Username: claims.Subject,
		Groups:   claims.Groups,
	}

	return nil
}
