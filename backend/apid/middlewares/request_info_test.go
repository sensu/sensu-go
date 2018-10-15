package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestRequestInfoVerb checks that the various HTTP methods are translated into
// the right verbs that Sensu uses internally, including "watch" and "list".
func TestRequestInfoVerb(t *testing.T) {
	cases := []struct {
		description string
		method      string
		expected    string
	}{
		{
			description: "HTTP POST is create",
			method:      "POST",
			expected:    "create",
		},
		{
			description: "HTTP GET is get",
			method:      "GET",
			expected:    "get",
		},
		{
			// This is not exactly correct, per HTTP HEAD semantics
			description: "HTTP HEAD is get",
			method:      "HEAD",
			expected:    "get",
		},
		{
			description: "HTTP PUT is update",
			method:      "PUT",
			expected:    "update",
		},
		{
			description: "HTTP DELETE is delete",
			method:      "DELETE",
			expected:    "delete",
		},
		{
			description: "invalid method gives empty verb",
			method:      "TEST",
			expected:    "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			checkResult := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				info := types.ContextRequestInfo(r.Context())
				assert.NotNil(t, info)
				assert.Equal(t, tt.expected, info.Verb)
			})

			middleware := RequestInfo{}
			server := httptest.NewServer(middleware.Then(checkResult))

			req, err := http.NewRequest(tt.method, server.URL, nil)
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
			expected:    types.RequestInfo{},
		},
		{
			description: "path not starting with API prefix",
			method:      "GET",
			path:        "/some/unrelated/path",
			expected:    types.RequestInfo{},
		},
		{
			description: "path only has API prefix",
			method:      "GET",
			path:        "/apis/",
			expected:    types.RequestInfo{},
		},
		{
			description: "path valid up to namespaces endpoint",
			method:      "GET",
			path:        "/apis/rbac/v1alpha1/namespaces",
			expected: types.RequestInfo{
				APIGroup:   "rbac",
				APIVersion: "v1alpha1",
				Verb:       "list",
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
				Verb:       "list",
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
