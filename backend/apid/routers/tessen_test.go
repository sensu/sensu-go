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
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/mock"
)

type mockTessenController struct {
	mock.Mock
}

func (m *mockTessenController) CreateOrUpdate(ctx context.Context, config *corev2.TessenConfig) error {
	return m.Called(ctx, config).Error(0)
}

func (m *mockTessenController) Get(ctx context.Context) (*corev2.TessenConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).(*corev2.TessenConfig), args.Error(1)
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
	b, _ := json.Marshal(corev2.DefaultTessenConfig())
	body := bytes.NewReader(b)
	endpoint := "/" + corev2.TessenResource
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

	fixture := corev2.DefaultTessenConfig()
	controller.On("Get", mock.Anything).Return(fixture, nil)
	endpoint := "/" + corev2.TessenResource
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

type mockTessenMetricController struct {
	mock.Mock
}

func (m *mockTessenMetricController) Publish(ctx context.Context, metrics []corev2.MetricPoint) error {
	return m.Called(ctx, metrics).Error(0)
}

func newTessenMetricTest(t *testing.T) (*mockTessenMetricController, *httptest.Server) {
	controller := &mockTessenMetricController{}
	tessenRouter := NewTessenMetricRouter(controller)
	router := mux.NewRouter()
	tessenRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPostTessenMetrics(t *testing.T) {
	controller, server := newTessenMetricTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("Publish", mock.Anything, mock.Anything).Return(nil)
	b, _ := json.Marshal([]corev2.MetricPoint{
		corev2.MetricPoint{
			Name:  "metric",
			Value: 1,
		},
	})
	body := bytes.NewReader(b)
	endpoint := "/api/core/v2/tessen/metrics"
	req := newRequest(t, http.MethodPost, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}
