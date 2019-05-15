package routers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

type mockVersionController struct {
	mock.Mock
}

func (m *mockVersionController) GetVersion(ctx context.Context) (*corev2.Version, error) {
	args := m.Called(ctx)
	return args.Get(0).(*corev2.Version), args.Error(1)
}

func newVersionTest(t *testing.T) (*mockVersionController, *httptest.Server) {
	controller := &mockVersionController{}
	versionRouter := NewVersionRouter(controller)
	router := mux.NewRouter()
	versionRouter.Mount(router)
	return controller, httptest.NewServer(router)
}
func TestVersionSuccess(t *testing.T) {
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

func TestVersionErr(t *testing.T) {
	controller, server := newVersionTest(t)
	defer server.Close()
	versionResponse := &corev2.Version{}
	controller.On("GetVersion", mock.Anything).Return(versionResponse, fmt.Errorf("foo"))

	client := new(http.Client)
	endpoint := "/version"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 500 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}
}
