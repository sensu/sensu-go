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

type mockOrgController struct {
	mock.Mock
}

func (m *mockOrgController) Create(ctx context.Context, org types.Namespace) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) CreateOrReplace(ctx context.Context, org types.Namespace) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) Update(ctx context.Context, org types.Namespace) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) Query(ctx context.Context) ([]*types.Namespace, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.Namespace), args.Error(1)
}

func (m *mockOrgController) Find(ctx context.Context, org string) (*types.Namespace, error) {
	args := m.Called(ctx, org)
	return args.Get(0).(*types.Namespace), args.Error(1)
}

func (m *mockOrgController) Destroy(ctx context.Context, org string) error {
	return m.Called(ctx, org).Error(0)
}

func newOrgTest(t *testing.T) (*mockOrgController, *httptest.Server) {
	controller := &mockOrgController{}
	orgRouter := NewNamespacesRouter(controller)
	router := mux.NewRouter()
	orgRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPostNamespace(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	org := types.FixtureNamespace("default")
	controller.On("Create", mock.Anything, mock.AnythingOfType("types.Namespace")).Return(nil)
	b, _ := json.Marshal(org)
	body := bytes.NewReader(b)
	endpoint := "/rbac/namespaces"
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
		mock.AnythingOfType("Namespace"))
}

func TestPutNamespace(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Namespace")).Return(nil)
	b, _ := json.Marshal(types.FixtureNamespace("default"))
	body := bytes.NewReader(b)
	endpoint := "/rbac/namespaces/default"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Namespace"))
}

func TestGetNamespace(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := types.FixtureNamespace("default")
	controller.On("Find", mock.Anything, "default").Return(fixture, nil)
	endpoint := "/rbac/namespaces/default"
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

func TestDeleteNamespace(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("Destroy", mock.Anything, "default").Return(nil)
	endpoint := "/rbac/namespaces/default"
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

func TestGetAllNamespaces(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	fixtures := []*types.Namespace{types.FixtureNamespace("default")}
	controller.On("Query", mock.Anything).Return(fixtures, nil)
	endpoint := "/rbac/namespaces"
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
