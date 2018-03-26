package dashboardd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardRouter(t *testing.T) {
	dashboard := Dashboardd{}
	router := httpRouter(&dashboard)

	testCases := []struct {
		path string
		want int
	}{
		{"/favicon.ico", http.StatusOK},
		{"/manifest.json", http.StatusOK},
		{"/asset-manifest.json", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("path %s", tc.path), func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)

			assert.Equal(t, tc.want, res.Code)
		})
	}
}
