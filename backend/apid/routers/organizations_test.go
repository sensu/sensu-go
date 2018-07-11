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

func (m *mockOrgController) Create(ctx context.Context, org types.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) CreateOrReplace(ctx context.Context, org types.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) Update(ctx context.Context, org types.Organization) error {
	return m.Called(ctx, org).Error(0)
}

func (m *mockOrgController) Query(ctx context.Context) ([]*types.Organization, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.Organization), args.Error(1)
}

func (m *mockOrgController) Find(ctx context.Context, org string) (*types.Organization, error) {
	args := m.Called(ctx, org)
	return args.Get(0).(*types.Organization), args.Error(1)
}

func (m *mockOrgController) Destroy(ctx context.Context, org string) error {
	return m.Called(ctx, org).Error(0)
}

func newOrgTest(t *testing.T) (*mockOrgController, *httptest.Server) {
	controller := &mockOrgController{}
	orgRouter := NewOrganizationsRouter(controller)
	router := mux.NewRouter()
	orgRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPostOrganization(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	org := types.FixtureOrganization("default")
	controller.On("Create", mock.Anything, mock.AnythingOfType("types.Organization")).Return(nil)
	b, _ := json.Marshal(org)
	body := bytes.NewReader(b)
	endpoint := "/rbac/organizations"
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
		mock.AnythingOfType("Organization"))
}

func TestPutOrganization(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Organization")).Return(nil)
	b, _ := json.Marshal(types.FixtureOrganization("default"))
	body := bytes.NewReader(b)
	endpoint := "/rbac/organizations/default"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "CreateOrReplace", mock.Anything, mock.AnythingOfType("types.Organization"))
}

func TestGetOrganization(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := types.FixtureOrganization("default")
	controller.On("Find", mock.Anything, "default").Return(fixture, nil)
	endpoint := "/rbac/organizations/default"
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

func TestDeleteOrganization(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("Destroy", mock.Anything, "default").Return(nil)
	endpoint := "/rbac/organizations/default"
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

func TestGetAllOrganizations(t *testing.T) {
	controller, server := newOrgTest(t)
	defer server.Close()

	client := new(http.Client)

	fixtures := []*types.Organization{types.FixtureOrganization("default")}
	controller.On("Query", mock.Anything).Return(fixtures, nil)
	endpoint := "/rbac/organizations"
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
