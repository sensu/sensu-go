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
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

type mockEnvController struct {
	mock.Mock
}

func (m *mockEnvController) Create(ctx context.Context, env types.Environment) error {
	return m.Called(ctx, env).Error(0)
}

func (m *mockEnvController) CreateOrReplace(ctx context.Context, env types.Environment) error {
	return m.Called(ctx, env).Error(0)
}

func (m *mockEnvController) Update(ctx context.Context, env types.Environment) error {
	return m.Called(ctx, env).Error(0)
}

func (m *mockEnvController) Query(ctx context.Context, org string) ([]*types.Environment, error) {
	args := m.Called(ctx, org)
	return args.Get(0).([]*types.Environment), args.Error(1)
}

func (m *mockEnvController) Find(ctx context.Context, org, env string) (*types.Environment, error) {
	args := m.Called(ctx, org, env)
	return args.Get(0).(*types.Environment), args.Error(1)
}

func (m *mockEnvController) Destroy(ctx context.Context, org, env string) error {
	return m.Called(ctx, org, env).Error(0)
}

func newEnvTest(t *testing.T) (*mockEnvController, *httptest.Server) {
	controller := &mockEnvController{}
	envRouter := NewEnvironmentsRouter(controller)
	router := mux.NewRouter()
	envRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPostEnvironment(t *testing.T) {
	controller, server := newEnvTest(t)
	defer server.Close()

	client := new(http.Client)

	env := types.FixtureEnvironment("default")
	controller.On("Create", mock.Anything, mock.AnythingOfType("types.Environment")).Return(nil)
	b, _ := json.Marshal(env)
	body := bytes.NewReader(b)
	endpoint := "/rbac/organizations/default/environments"
	req := newRequest(t, http.MethodPost, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t,
		"Create",
		mock.Anything,
		mock.AnythingOfType("Environment"))
}

func TestPutEnvironment(t *testing.T) {
	controller, server := newEnvTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Environment")).Return(nil)
	b, _ := json.Marshal(types.FixtureEnvironment("default"))
	body := bytes.NewReader(b)
	endpoint := "/rbac/organizations/default/environments/default"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Environment"))
}

func TestGetEnvironment(t *testing.T) {
	controller, server := newEnvTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := types.FixtureEnvironment("default")
	controller.On("Find", mock.Anything, "default", "default").Return(fixture, nil)
	endpoint := "/rbac/organizations/default/environments/default"
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

func TestDeleteEnvironment(t *testing.T) {
	controller, server := newEnvTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("Destroy", mock.Anything, "default", "default").Return(nil)
	endpoint := "/rbac/organizations/default/environments/default"
	req := newRequest(t, http.MethodDelete, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}
}

func TestGetAllEnvironments(t *testing.T) {
	controller, server := newEnvTest(t)
	defer server.Close()

	client := new(http.Client)

	fixtures := []*types.Environment{types.FixtureEnvironment("default")}
	controller.On("Query", mock.Anything, "default").Return(fixtures, nil)
	endpoint := "/rbac/organizations/default/environments"
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
