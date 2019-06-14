package routers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

type mockClusterIDController struct {
	mock.Mock
}

func (m *mockClusterIDController) Get(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}

func newClusterIDTest(t *testing.T) (*mockClusterIDController, *httptest.Server) {
	controller := &mockClusterIDController{}
	clusterIDRouter := NewClusterIDRouter(controller)
	router := mux.NewRouter()
	clusterIDRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestGetClusterID(t *testing.T) {
	controller, server := newClusterIDTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := uuid.New().String()
	controller.On("Get", mock.Anything).Return(fixture, nil)
	endpoint := "/id"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "Get", mock.Anything)
}
