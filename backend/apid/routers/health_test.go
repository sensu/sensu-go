package routers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

type mockHealthController struct {
	mock.Mock
}

func (m *mockHealthController) GetClusterHealth(ctx context.Context) *types.HealthResponse {
	args := m.Called(ctx)
	return args.Get(0).(*types.HealthResponse)
}

func newHealthTest(t *testing.T) (*mockHealthController, *httptest.Server) {
	controller := &mockHealthController{}
	healthRouter := NewHealthRouter(controller)
	router := mux.NewRouter()
	healthRouter.Mount(router)
	return controller, httptest.NewServer(router)
}
func TestHealthSuccess(t *testing.T) {
	controller, server := newHealthTest(t)
	defer server.Close()
	healthResponse := &types.HealthResponse{}
	controller.On("GetClusterHealth", mock.Anything).Return(healthResponse)

	client := new(http.Client)
	endpoint := "/health"
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

func TestHealthyCluster(t *testing.T) {
	controller, server := newHealthTest(t)
	defer server.Close()
	healthResponse := types.FixtureHealthResponse(true)
	controller.On("GetClusterHealth", mock.Anything).Return(healthResponse)

	client := new(http.Client)
	endpoint := "/health"
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

func TestUnHealthyClusterStatus(t *testing.T) {
	controller, server := newHealthTest(t)
	defer server.Close()
	healthResponse := types.FixtureHealthResponse(false)
	controller.On("GetClusterHealth", mock.Anything).Return(healthResponse)

	client := new(http.Client)
	endpoint := "/health"
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
