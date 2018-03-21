package dashboardd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDashboardRouter(t *testing.T) {
	dashboard := Dashboardd{}
	router := httpRouter(&dashboard)

	testCases := []struct {
		path string
		want int
	}{
		{"/auth", http.StatusOK},
		{"/graphql", http.StatusOK},
		{"/test", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path %s", tc.path), func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			t.Skip("What do we want to assert here")
		})
	}
}
