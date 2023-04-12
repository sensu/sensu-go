package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceMiddlware(t *testing.T) {
	cases := []struct {
		description	string
		method		string
		url		string
		urlVars		map[string]string
		expected	string
	}{
		{
			description:	"No query param or path variable",
			method:		"GET",
			url:		"/",
			expected:	"",
		},
		{
			description:	"Path variable",
			method:		"GET",
			url:		"/",
			urlVars: map[string]string{
				"namespace": "foobar",
			},
			expected:	"foobar",
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var namespace string
				if value := r.Context().Value(v2.NamespaceKey); value != nil {
					namespace = value.(string)
				} else {
					namespace = ""
				}
				assert.Equal(t, tt.expected, namespace)
			})
			middleware := Namespace{}

			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.method, tt.url, nil)
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
