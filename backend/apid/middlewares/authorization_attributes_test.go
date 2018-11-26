package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	sensuJWT "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

func TestAuthorizationAttributes(t *testing.T) {
	cases := []struct {
		description string
		method      string
		path        string
		expected    authorization.Attributes
	}{
		{
			description: "GET /",
			method:      "GET",
			path:        "/",
			expected:    authorization.Attributes{Verb: "list"},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces",
			method:      "GET",
			path:        "/apis/core/v1alpha1/namespaces",
			expected: authorization.Attributes{
				APIGroup:   "core",
				APIVersion: "v1alpha1",
				Resource:   "namespaces",
				Verb:       "list",
			},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces/default",
			method:      "GET",
			path:        "/apis/core/v1alpha1/namespaces/default",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v1alpha1",
				Resource:     "namespaces",
				ResourceName: "default",
				Verb:         "get",
			},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces/default/checks",
			method:      "GET",
			path:        "/apis/core/v1alpha1/namespaces/default/checks",
			expected: authorization.Attributes{
				APIGroup:   "core",
				APIVersion: "v1alpha1",
				Namespace:  "default",
				Resource:   "checks",
				Verb:       "list",
			},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces/default/checks/check-cpu",
			method:      "GET",
			path:        "/apis/core/v1alpha1/namespaces/default/checks/check-cpu",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v1alpha1",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "check-cpu",
				Verb:         "get",
			},
		},
		{
			description: "DELETE /foo",
			method:      "DELETE",
			path:        "/foo",
			expected: authorization.Attributes{
				Verb: "delete",
			},
		},
		{
			description: "POST /foo",
			method:      "POST",
			path:        "/foo",
			expected: authorization.Attributes{
				Verb: "create",
			},
		},
		{
			description: "PUT /foo",
			method:      "PUT",
			path:        "/foo",
			expected: authorization.Attributes{
				Verb: "update",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attrs := authorization.GetAttributes(r.Context())
				assert.NotNil(t, attrs)

				// Inject our user in the expected attributes
				tt.expected.User = types.User{Username: "admin"}

				assert.Equal(t, &tt.expected, attrs)
			})

			w := httptest.NewRecorder()

			// Prepare the request
			r, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}
			claims := types.Claims{StandardClaims: jwt.StandardClaims{Subject: "admin"}}
			ctx := sensuJWT.SetClaimsIntoContext(r, &claims)

			// Prepare the router
			router := mux.NewRouter()
			router.PathPrefix("/apis/{group}/{version}/namespaces/{namespace}/{resource}/{name}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/namespaces/{namespace}/{resource}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/{resource}/{name}").Handler(testHandler)
			router.PathPrefix("/apis/{group}/{version}/{resource}").Handler(testHandler)
			router.PathPrefix("/").Handler(testHandler) // catch all
			middleware := AuthorizationAttributes{}
			router.Use(middleware.Then)

			// Serve the request
			router.ServeHTTP(w, r.WithContext(ctx))
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestLegacyAuthorizationAttributes(t *testing.T) {
	cases := []struct {
		description string
		method      string
		path        string
		expected    authorization.Attributes
	}{
		{
			description: "GET /assets",
			method:      "GET",
			path:        "/assets",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "assets",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /assets?namespace=foo",
			method:      "GET",
			path:        "/assets?namespace=foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "foo",
				Resource:     "assets",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /assets/foo",
			method:      "GET",
			path:        "/assets/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "assets",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /assets/foo?namespace=bar",
			method:      "GET",
			path:        "/assets/foo?namespace=bar",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "bar",
				Resource:     "assets",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /checks",
			method:      "GET",
			path:        "/checks",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /checks/foo",
			method:      "GET",
			path:        "/checks/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "POST /checks/foo/execute",
			method:      "POST",
			path:        "/checks/foo/execute",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "create",
			},
		},
		{
			description: "PUT /checks/foo/hooks/bar",
			method:      "PUT",
			path:        "/checks/foo/hooks/bar",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "update",
			},
		},
		{
			description: "DELETE /checks/foo/hooks/bar/hook/baz",
			method:      "DELETE",
			path:        "/checks/foo/hooks/bar/hook/baz",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "delete",
			},
		},
		{
			description: "GET /cluster/members",
			method:      "GET",
			path:        "/cluster/members",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "cluster-members",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /cluster/members/foo",
			method:      "GET",
			path:        "/cluster/members/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "cluster-members",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /events",
			method:      "GET",
			path:        "/events",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "events",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /events/entity_name",
			method:      "GET",
			path:        "/events/entity_name",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "events",
				ResourceName: "entity_name",
				Verb:         "list",
			},
		},
		{
			description: "GET /events/entity_name/check_name",
			method:      "GET",
			path:        "/events/entity_name/check_name",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "events",
				ResourceName: "entity_name/check_name",
				Verb:         "get",
			},
		},
		{
			description: "GET /rbac/namespaces",
			method:      "GET",
			path:        "/rbac/namespaces",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "namespaces",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /rbac/namespaces/foo",
			method:      "GET",
			path:        "/rbac/namespaces/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "namespaces",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /rbac/users",
			method:      "GET",
			path:        "/rbac/users",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "users",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /rbac/users/foo",
			method:      "GET",
			path:        "/rbac/users/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /silenced",
			method:      "GET",
			path:        "/silenced",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /silenced/foo",
			method:      "GET",
			path:        "/silenced/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /silenced/checks",
			method:      "GET",
			path:        "/silenced/checks",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "checks",
				Verb:         "list",
			},
		},
		{
			description: "GET /silenced/checks/foo",
			method:      "GET",
			path:        "/silenced/checks/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "checks/foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /silenced/subscriptions",
			method:      "GET",
			path:        "/silenced/subscriptions",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "subscriptions",
				Verb:         "list",
			},
		},
		{
			description: "GET /silenced/subscriptions/foo",
			method:      "GET",
			path:        "/silenced/subscriptions/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced",
				ResourceName: "subscriptions/foo",
				Verb:         "get",
			},
		},
		{
			description: "View another user",
			method:      "GET",
			path:        "/rbac/users/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "View itself",
			method:      "GET",
			path:        "/rbac/users/admin",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     types.LocalSelfUserResource,
				ResourceName: "admin",
				Verb:         "get",
			},
		},
		{
			description: "Update another user password",
			method:      "PUT",
			path:        "/rbac/users/foo/password",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "update",
			},
		},
		{
			description: "Update its own password",
			method:      "PUT",
			path:        "/rbac/users/admin/password",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     types.LocalSelfUserResource,
				ResourceName: "admin",
				Verb:         "update",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attrs := authorization.GetAttributes(r.Context())
				assert.NotNil(t, attrs)

				// Inject our user in the expected attributes
				tt.expected.User = types.User{Username: "admin"}

				assert.Equal(t, &tt.expected, attrs)
			})

			// Prepare our HTTP server
			w := httptest.NewRecorder()

			// Prepare our request
			r, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}
			claims := types.Claims{StandardClaims: jwt.StandardClaims{Subject: "admin"}}
			ctx := sensuJWT.SetClaimsIntoContext(r, &claims)

			// Prepare the router
			middleware := LegacyAuthorizationAttributes{}
			router := mux.NewRouter()
			router.PathPrefix("/{resource:events}/{entity}/{check}").Handler(testHandler)
			router.PathPrefix("/{resource:events}/{entity}").Handler(testHandler)
			router.PathPrefix("/{resource:silenced}/checks/{check}").Handler(testHandler)
			router.PathPrefix("/{resource:silenced}/subscriptions/{subscription}").Handler(testHandler)
			router.PathPrefix("/{prefix:cluster}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/{prefix:cluster}/{resource}").Handler(testHandler)
			router.PathPrefix("/{prefix:rbac}/{resource}/{id}/{subresource}").Handler(testHandler)
			router.PathPrefix("/{prefix:rbac}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/{prefix:rbac}/{resource}").Handler(testHandler)
			router.PathPrefix("/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/").Handler(testHandler) // catch all for legacy routes
			router.Use(middleware.Then)

			// Serve the request
			router.ServeHTTP(w, r.WithContext(ctx))

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
