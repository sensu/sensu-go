package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestInfo(t *testing.T) {
	cases := []struct {
		description string
		method      string
		urlVars     map[string]string
		expected    types.RequestInfo
	}{
		{
			description: "GET /",
			method:      "GET",
			urlVars:     map[string]string{},
			expected:    types.RequestInfo{Verb: "list"},
		},
		{
			description: "GET /apis/core/v1alpha1/namespaces",
			method:      "GET",
			urlVars: map[string]string{
				"group":   "core",
				"version": "v1alpha1",
				"kind":    "namespaces",
			},
			expected: types.RequestInfo{
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
			expected: types.RequestInfo{
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
			expected: types.RequestInfo{
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
			expected: types.RequestInfo{
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
			expected: types.RequestInfo{
				Verb: "delete",
			},
		},
		{
			description: "POST /foo",
			method:      "POST",
			urlVars:     map[string]string{},
			expected: types.RequestInfo{
				Verb: "create",
			},
		},
		{
			description: "PUT /foo",
			method:      "PUT",
			urlVars:     map[string]string{},
			expected: types.RequestInfo{
				Verb: "update",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				info := types.ContextRequestInfo(r.Context())
				assert.NotNil(t, info)
				assert.Equal(t, &tt.expected, info)
			})
			middleware := RequestInfo{}

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
