package routers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
)

func TestClusterRoleBindingsRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	router := NewClusterRoleBindingsRouter(s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)

	pathPrefix := "/clusterrolebindings"
	kind := "*v2.ClusterRoleBinding"
	fixture := corev2.FixtureClusterRoleBinding("foo")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(pathPrefix, kind, fixture)...)
	tests = append(tests, listTestCases(pathPrefix, kind, []corev2.Resource{fixture})...)
	tests = append(tests, createTestCases(pathPrefix, kind)...)
	tests = append(tests, updateTestCases(pathPrefix, kind)...)
	tests = append(tests, deleteTestCases(pathPrefix, kind)...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only start the HTTP server here to prevent data races in tests
			server := httptest.NewServer(parentRouter)
			defer server.Close()

			if tt.storeFunc != nil {
				tt.storeFunc(s)
			}

			// Prepare the HTTP request
			client := new(http.Client)
			req, err := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Perform the HTTP request
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// Inspect the response code
			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, wantStatusCode %v", res.StatusCode, tt.wantStatusCode)
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("error message: %q", string(body))
				return
			}
		})
	}
}
