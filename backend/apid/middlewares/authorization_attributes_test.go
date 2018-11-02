package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationAttributes(t *testing.T) {
	cases := []struct {
		description string
		method      string
		urlVars     map[string]string
		expected    authorization.Attributes
	}{
		{
			description: "GET /",
			method:      "GET",
			urlVars:     map[string]string{},
			expected:    authorization.Attributes{Verb: "list"},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces",
			method:      "GET",
			urlVars: map[string]string{
				"group":   "core",
				"version": "v1alpha1",
				"kind":    "namespaces",
			},
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
			urlVars: map[string]string{
				"group":   "core",
				"version": "v1alpha1",
				"kind":    "namespaces",
				"name":    "default",
			},
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
			urlVars: map[string]string{
				"group":     "core",
				"version":   "v1alpha1",
				"namespace": "default",
				"kind":      "checks",
			},
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
			urlVars: map[string]string{
				"group":     "core",
				"version":   "v1alpha1",
				"namespace": "default",
				"kind":      "checks",
				"name":      "check-cpu",
			},
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
			urlVars:     map[string]string{},
			expected: authorization.Attributes{
				Verb: "delete",
			},
		},
		{
			description: "POST /foo",
			method:      "POST",
			urlVars:     map[string]string{},
			expected: authorization.Attributes{
				Verb: "create",
			},
		},
		{
			description: "PUT /foo",
			method:      "PUT",
			urlVars:     map[string]string{},
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
				assert.Equal(t, &tt.expected, attrs)
			})
			middleware := AuthorizationAttributes{}

			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			r = mux.SetURLVars(r, tt.urlVars)
			handler := middleware.Then(testHandler)
			handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestLegacyAuthorizationAttributes(t *testing.T) {
	cases := []struct {
		description string
		method      string
		path        string
		urlVars     map[string]string
		expected    authorization.Attributes
	}{
		{
			description: "GET /assets",
			method:      "GET",
			path:        "/assets",
			urlVars:     map[string]string{},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars: map[string]string{
				"id": "foo",
			},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "post",
			},
		},
		{
			description: "PUT /checks/foo/hooks/bar",
			method:      "PUT",
			path:        "/checks/foo/hooks/bar",
			urlVars: map[string]string{
				"id":   "foo",
				"type": "bar",
			},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "checks",
				ResourceName: "foo",
				Verb:         "put",
			},
		},
		{
			description: "DELETE /checks/foo/hooks/bar/hook/baz",
			method:      "DELETE",
			path:        "/checks/foo/hooks/bar/hook/baz",
			urlVars: map[string]string{
				"id":   "foo",
				"type": "bar",
				"hook": "baz",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"entity": "entity_name",
			},
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
			urlVars: map[string]string{
				"check":  "check_name",
				"entity": "entity_name",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars:     map[string]string{},
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
			urlVars: map[string]string{
				"id": "foo",
			},
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
			urlVars:     map[string]string{},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced/checks",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /silenced/checks/foo",
			method:      "GET",
			path:        "/silenced/checks/foo",
			urlVars: map[string]string{
				"check": "foo",
			},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced/checks",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
		{
			description: "GET /silenced/subscriptions",
			method:      "GET",
			path:        "/silenced/subscriptions",
			urlVars:     map[string]string{},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced/subscriptions",
				ResourceName: "",
				Verb:         "list",
			},
		},
		{
			description: "GET /silenced/subscriptions/foo",
			method:      "GET",
			path:        "/silenced/subscriptions/foo",
			urlVars: map[string]string{
				"subscription": "foo",
			},
			expected: authorization.Attributes{
				APIGroup:     "core",
				APIVersion:   "v2",
				Namespace:    "default",
				Resource:     "silenced/subscriptions",
				ResourceName: "foo",
				Verb:         "get",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attrs := authorization.GetAttributes(r.Context())
				assert.NotNil(t, attrs)
				assert.Equal(t, &tt.expected, attrs)
			})
			middleware := LegacyAuthorizationAttributes{}

			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			r = mux.SetURLVars(r, tt.urlVars)
			handler := middleware.Then(testHandler)
			handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
