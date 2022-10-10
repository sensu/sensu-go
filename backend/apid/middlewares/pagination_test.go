package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	corev2 "github.com/sensu/core/v2"
)

func TestPaginationMiddleware(t *testing.T) {
	cases := []struct {
		description           string
		queryParams           string
		expectedLimit         int
		expectedContinueToken string
	}{
		{
			description:           "No query parameters",
			queryParams:           "",
			expectedLimit:         0,
			expectedContinueToken: "",
		},
		{
			description:           "Only limit used",
			queryParams:           "?limit=100",
			expectedLimit:         100,
			expectedContinueToken: "",
		},
		{
			description:           "Only continue used",
			queryParams:           "?continue=Y2FtdXM",
			expectedLimit:         0,
			expectedContinueToken: "camus",
		},
		{
			description:           "Both limit and continue used",
			queryParams:           "?limit=42&continue=c2FydHJl",
			expectedLimit:         42,
			expectedContinueToken: "sartre",
		},
		{
			description:           "Invalid limit",
			queryParams:           "?limit=sandwich",
			expectedLimit:         0,
			expectedContinueToken: "",
		},
		{
			description:           "Invalid continue",
			queryParams:           "?continue=cake%QQ",
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
			r, err := http.NewRequest("GET", "/"+tt.queryParams, nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			handler := middleware.Then(testHandler)
			handler.ServeHTTP(w, r)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
