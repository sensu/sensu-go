package routers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/mock"
)

type mockVersionController struct {
	mock.Mock
}

func (m *mockVersionController) GetVersion(ctx context.Context) *corev2.Version {
	args := m.Called(ctx)
	return args.Get(0).(*corev2.Version)
}

func newVersionTest(t *testing.T) (*mockVersionController, *httptest.Server) {
	controller := &mockVersionController{}
	versionRouter := NewVersionRouter(controller)
	router := mux.NewRouter()
	versionRouter.Mount(router)
	return controller, httptest.NewServer(router)
}
func TestVersion(t *testing.T) {
	controller, server := newVersionTest(t)
	defer server.Close()
	versionResponse := corev2.FixtureVersion()
	controller.On("GetVersion", mock.Anything).Return(versionResponse, nil)

	client := new(http.Client)
	endpoint := "/version"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}
}
