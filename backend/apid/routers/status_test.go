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

func passStatus() func() types.StatusMap {
	return func() types.StatusMap {
		status := map[string]bool{
			"serviceAlive": true,
		}
		return status
	}
}

func failStatus() func() types.StatusMap {
	return func() types.StatusMap {
		status := map[string]bool{
			"serviceAlive": false,
		}
		return status
	}
}

type mockHealthController struct {
	mock.Mock
}

func (m *mockHealthController) GetClusterHealth(ctx context.Context) []*types.ClusterHealth {
	args := m.Called(ctx)
	return args.Get(0).([]*types.ClusterHealth)
}

func newStatusTest(t *testing.T, fn func() types.StatusMap) (*mockHealthController, *httptest.Server) {
	controller := &mockHealthController{}
	statusRouter := NewStatusRouter(fn, controller)
	router := mux.NewRouter()
	statusRouter.Mount(router)
	return controller, httptest.NewServer(router)
}

func TestStatusInfo(t *testing.T) {
	controller, server := newStatusTest(t, passStatus())
	defer server.Close()
	controller.On("GetClusterHealth", mock.Anything).Return([]*types.ClusterHealth{})
	client := new(http.Client)
	endpoint := "/info"
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

func TestHealthStatusSuccess(t *testing.T) {
	controller, server := newStatusTest(t, passStatus())
	defer server.Close()
	controller.On("GetClusterHealth", mock.Anything).Return([]*types.ClusterHealth{})

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

func TestHealthStatusFail(t *testing.T) {
	controller, server := newStatusTest(t, failStatus())
	defer server.Close()
	controller.On("GetClusterHealth", mock.Anything).Return([]*types.ClusterHealth{})

	client := new(http.Client)
	endpoint := "/health"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode <= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("expected bad status, got: %d (%q)", resp.StatusCode, string(body))
	}
}

func TestHealthyClusterStatus(t *testing.T) {
	controller, server := newStatusTest(t, passStatus())
	defer server.Close()
	clusterHealth := []*types.ClusterHealth{}
	clusterHealth = append(clusterHealth, &types.ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend0",
		Err:      nil,
		Healthy:  true,
	})
	controller.On("GetClusterHealth", mock.Anything).Return(clusterHealth)

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
