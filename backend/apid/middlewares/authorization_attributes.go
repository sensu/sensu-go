package middlewares

import (
	"context"
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
// /apis/{group}/{version}/namespaces
// /apis/{group}/{version}/namespaces/{namespace}/{resource}
// /apis/{group}/{version}/namespaces/{namespace}/{resource}/{name}
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
		case http.MethodPost:
			attrs.Verb = "create"
		case http.MethodGet, http.MethodHead:
			attrs.Verb = "get"
		case http.MethodPatch, http.MethodPut:
			attrs.Verb = "update"
		case http.MethodDelete:
			attrs.Verb = "delete"
		default:
			attrs.Verb = ""
		}

		vars := mux.Vars(r)
		attrs.APIGroup = vars["group"]
		attrs.APIVersion = vars["version"]
		attrs.Namespace = vars["namespace"]
		attrs.Resource = vars["resource"]
		attrs.ResourceName = vars["id"]

		// TODO: we can probably get rid of this special case by reworking the
		// cluster router.
		if attrs.Resource == "cluster" {
			attrs.Resource = "cluster-members"
		}

		// Most resource names are identified by a route variable named "id".
		// Other resources have snowflake paths; see their corresponding router
		// and the expected paths above.
		switch attrs.Resource {
		case "events":
			attrs.ResourceName = path.Join(vars["entity"], vars["check"])
		case "silenced":
			if strings.Contains(r.URL.Path, "/silenced/checks") {
				attrs.ResourceName = path.Join("checks", vars["check"])
			} else if strings.Contains(r.URL.Path, "/silenced/subscriptions") {
				attrs.ResourceName = path.Join("subscriptions", vars["subscription"])
			}
		}

		if attrs.Verb == "get" && (attrs.ResourceName == "" || isListable(attrs.Resource, attrs.ResourceName)) {
			attrs.Verb = "list"
		}

		// Add the user to the attributes
		if err := GetUser(ctx, attrs); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Verify if the authenticated user is trying to access itself
		if attrs.Resource == "users" && attrs.ResourceName == attrs.User.Username {
			// Change the resource to LocalSelfUserResource if a user views itself
			if attrs.Verb == "get" && vars["subresource"] == "" {
				attrs.Resource = types.LocalSelfUserResource
			}

			switch vars["subresource"] {
			case "password", "change_password":
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

// GetUser retrieves the user from the context and inject it into the
// authorization attributes
func GetUser(ctx context.Context, attrs *authorization.Attributes) error {
	// Get the claims from the request context
	claims := jwt.GetClaimsFromContext(ctx)
	if claims == nil {
		return authorization.ErrNoClaims
	}

	// Add the user to our request info
	attrs.User = types.User{
		Username: claims.Subject,
		Groups:   claims.Groups,
	}

	return nil
}
