package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestInfoVerb(t *testing.T) {
	// TODO: Test focused on the translation of HTTP methods and "special verbs",
	// such as "watch" and "list", to a Verb by the middleware.
}

func TestRequestInfo(t *testing.T) {
	cases := []struct {
		description string
		method      string
		path        string
		expected    types.RequestInfo
	}{
		{
			description: "path is empty",
			method:      "GET",
			path:        "",
			expected: types.RequestInfo{
				Verb: "get",
			},
		},
		{
			description: "path not starting with API prefix",
			method:      "GET",
			path:        "/some/unrelated/path",
			expected: types.RequestInfo{
				Verb: "get",
			},
		},
		{
			description: "path only has API prefix",
			method:      "GET",
			path:        "/apis/",
			expected: types.RequestInfo{
				Verb: "get",
			},
		},
		{
			description: "path valid up to API group",
			method:      "GET",
			path:        "/apis/rbac/",
			expected: types.RequestInfo{
				APIGroup: "rbac",
				Verb:     "get",
			},
		},
		{
			description: "path valid up to API version",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Verb:       "get",
			},
		},
		{
			description: "path with wrong namespace prefix",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/some/unrelated/path",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Verb:       "get",
			},
		},
		{
			description: "path valid up to namespaces endpoint",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Verb:       "get",
			},
		},
		{
			description: "path valid up to specific namespace",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces/my-namespace",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Namespace:  "my-namespace",
				Verb:       "get",
			},
		},
		{
			description: "path valid up to specific resource type",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces/my-namespace/check",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Namespace:  "my-namespace",
				Resource:   "check",
				Verb:       "get",
			},
		},
		{
			description: "path valid up to specific resource name",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces/my-namespace/check/my-check",
			expected: types.RequestInfo{
				APIGroup:     "rbac",
				APIVersion:   "v1alpha1",
				Namespace:    "my-namespace",
				Resource:     "check",
				ResourceName: "my-check",
				Verb:         "get",
			},
		},
		{
			description: "ignore path beyond resource name",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces/my-namespace/check/my-check/ignored/bits",
			expected: types.RequestInfo{
				APIGroup:     "rbac",
				APIVersion:   "v1alpha1",
				Namespace:    "my-namespace",
				Resource:     "check",
				ResourceName: "my-check",
				Verb:         "get",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			checkResult := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				info := types.ContextRequestInfo(r.Context())
				assert.NotNil(t, info)
				assert.Equal(t, &tt.expected, info)
			})

			middleware := RequestInfo{}
			server := httptest.NewServer(middleware.Then(checkResult))

			req, err := http.NewRequest(tt.method, server.URL+tt.path, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal("Failed to get a response: ", err)
			}

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			server.Close()
		})
	}
}
