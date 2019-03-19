package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestPaginationMiddleware(t *testing.T) {
	cases := []struct {
		description           string
		urlVars               map[string]string
		expectedLimit         int
		expectedContinueToken string
	}{
		{
			description:           "No query parameters",
			urlVars:               map[string]string{},
			expectedLimit:         0,
			expectedContinueToken: "",
		},
		{
			description: "Only limit used",
			urlVars: map[string]string{
				"limit": "100",
			},
			expectedLimit:         100,
			expectedContinueToken: "",
		},
		{
			description: "Only continue used",
			urlVars: map[string]string{
				"continue": "camus",
			},
			expectedLimit:         0,
			expectedContinueToken: "camus",
		},
		{
			description: "Both limit and continue used",
			urlVars: map[string]string{
				"limit":    "42",
				"continue": "sartre",
			},
			expectedLimit:         42,
			expectedContinueToken: "sartre",
		},
		{
			description: "Invalid limit",
			urlVars: map[string]string{
				"limit": "sandwich",
			},
			expectedLimit:         0,
			expectedContinueToken: "",
		},
		{
			description: "Invalid continue",
			urlVars: map[string]string{
				"continue": "cake%QQ",
			},
			expectedLimit:         0,
			expectedContinueToken: "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var limit int
				var continueToken string

				if value := r.Context().Value(corev2.PageSizeKey); value != nil {
					limit = value.(int)
				}
				assert.Equal(t, tt.expectedLimit, limit)

				if value := r.Context().Value(corev2.PageContinueKey); value != nil {
					continueToken = value.(string)
				}
				assert.Equal(t, tt.expectedContinueToken, continueToken)
			})

			middleware := Pagination{}

			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/", nil)
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
