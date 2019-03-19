package routers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/tessen"
	"github.com/stretchr/testify/mock"
)

type mockTessenController struct {
	mock.Mock
}

func (m *mockTessenController) CreateOrUpdate(ctx context.Context, config *tessen.Config) error {
	return m.Called(ctx, config).Error(0)
}

func (m *mockTessenController) Get(ctx context.Context) (*tessen.Config, error) {
	args := m.Called(ctx)
	return args.Get(0).(*tessen.Config), args.Error(1)
}

func newTessenTest(t *testing.T) (*mockTessenController, *httptest.Server) {
	controller := &mockTessenController{}
	tessenRouter := NewTessenRouter(controller)
	router := mux.NewRouter()
	tessenRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPutTessen(t *testing.T) {
	controller, server := newTessenTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	b, _ := json.Marshal(tessen.DefaultConfig())
	body := bytes.NewReader(b)
	endpoint := "/tessen"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything)
}

func TestGetTessen(t *testing.T) {
	controller, server := newTessenTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := tessen.DefaultConfig()
	controller.On("Get", mock.Anything).Return(fixture, nil)
	endpoint := "/tessen"
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
