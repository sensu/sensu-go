package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt"
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
			description: "GET /api/core/v1alpha1/namespaces",
			method:      "GET",
			path:        "/api/core/v1alpha1/namespaces",
			expected: authorization.Attributes{
				APIGroup:   "core",
				APIVersion: "v1alpha1",
				Resource:   "namespaces",
				Verb:       "list",
			},
		},
		{
			description: "GET /api/core/v1alpha1/namespaces/default",
			method:      "GET",
			path:        "/api/core/v1alpha1/namespaces/default",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v1alpha1",
				Resource:     "namespaces",
				ResourceName: "default",
				Verb:         "get",
			},
		},
		{
			description: "GET /api/core/v1alpha1/namespaces/default/checks",
			method:      "GET",
			path:        "/api/core/v1alpha1/namespaces/default/checks",
			expected: authorization.Attributes{
				APIGroup:   "core",
				APIVersion: "v1alpha1",
				Namespace:  "default",
				Resource:   "checks",
				Verb:       "list",
			},
		},
		{
			description: "GET /api/core/v1alpha1/namespaces/default/checks/check-cpu",
			method:      "GET",
			path:        "/api/core/v1alpha1/namespaces/default/checks/check-cpu",
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
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/{resource}").Handler(testHandler)
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
			description: "GET /api/core/v2/namespaces/default/assets",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/assets",
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
			description: "GET /api/core/v2/namespaces/foo/assets",
			method:      "GET",
			path:        "/api/core/v2/namespaces/foo/assets",
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
			description: "GET /api/core/v2/namespaces/default/assets/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/assets/foo",
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
			description: "GET /api/core/v2/namespaces/bar/assets/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/bar/assets/foo",
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
			description: "GET /api/core/v2/namespaces/default/checks",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/checks",
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
			description: "GET /api/core/v2/namespaces/default/checks/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/checks/foo",
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
			description: "POST /api/core/v2/namespaces/default/checks/foo/execute",
			method:      "POST",
			path:        "/api/core/v2/namespaces/default/checks/foo/execute",
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
			description: "PUT /api/core/v2/namespaces/default/checks/foo/hooks/bar",
			method:      "PUT",
			path:        "/api/core/v2/namespaces/default/checks/foo/hooks/bar",
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
			description: "DELETE /api/core/v2/namespaces/default/checks/foo/hooks/bar/hook/baz",
			method:      "DELETE",
			path:        "/api/core/v2/namespaces/default/checks/foo/hooks/bar/hook/baz",
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
			description: "GET /api/core/v2/namespaces/default/cluster/members",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/cluster/members",
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
			description: "GET /api/core/v2/namespaces/default/cluster/members/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/cluster/members/foo",
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
			description: "GET /api/core/v2/namespaces/default/events",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/events",
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
			description: "GET /api/core/v2/namespaces/default/events/entity_name",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/events/entity_name",
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
			description: "GET /api/core/v2/namespaces/default/events/entity_name/check_name",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/events/entity_name/check_name",
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
			description: "GET /api/core/v2/namespaces",
			method:      "GET",
			path:        "/api/core/v2/namespaces",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Resource:     "namespaces",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /api/core/v2/namespaces/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Resource:     "namespaces",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /api/core/v2/users",
			method:      "GET",
			path:        "/api/core/v2/users",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     "users",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /api/core/v2/users/foo",
			method:      "GET",
			path:        "/api/core/v2/users/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /api/core/v2/namespaces/default/silenced",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced",
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
			description: "GET /api/core/v2/namespaces/default/silenced/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced/foo",
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
			description: "GET /api/core/v2/namespaces/default/silenced/checks",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced/checks",
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
			description: "GET /api/core/v2/namespaces/default/silenced/checks/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced/checks/foo",
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
			description: "GET /api/core/v2/namespaces/default/silenced/subscriptions",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced/subscriptions",
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
			description: "GET /api/core/v2/namespaces/default/silenced/subscriptions/foo",
			method:      "GET",
			path:        "/api/core/v2/namespaces/default/silenced/subscriptions/foo",
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
			path:        "/api/core/v2/users/foo",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "View itself",
			method:      "GET",
			path:        "/api/core/v2/users/admin",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     types.LocalSelfUserResource,
				ResourceName: "admin",
				Verb:         "get",
			},
		},
		{
			description: "Update another user password",
			method:      "PUT",
			path:        "/api/core/v2/users/foo/password",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
				Resource:     "users",
				ResourceName: "foo",
				Verb:         "update",
			},
		},
		{
			description: "Update its own password",
			method:      "PUT",
			path:        "/api/core/v2/users/admin/password",
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "",
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
			middleware := AuthorizationAttributes{}
			router := mux.NewRouter()
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:cluster}/members/{id}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:cluster}/members").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:events}/{entity}/{check}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:events}/{entity}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:silenced}/checks/{check}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource:silenced}/subscriptions/{subscription}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/namespaces/{namespace}/{resource}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/{resource}/{id}/{subresource}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/{resource}/{id}").Handler(testHandler)
			router.PathPrefix("/api/{group}/{version}/{resource}").Handler(testHandler)
			router.PathPrefix("/").Handler(testHandler) // catch all
			router.Use(middleware.Then)

			// Serve the request
			router.ServeHTTP(w, r.WithContext(ctx))

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
