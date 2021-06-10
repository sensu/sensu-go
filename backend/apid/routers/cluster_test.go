package routers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/etcd/client/v3"
)

type mockClusterController struct {
	mock.Mock
}

func (m *mockClusterController) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*clientv3.MemberListResponse), args.Error(1)
}

func (m *mockClusterController) MemberAdd(ctx context.Context, peerAddrs []string) (*clientv3.MemberAddResponse, error) {
	args := m.Called(ctx, peerAddrs)
	return args.Get(0).(*clientv3.MemberAddResponse), args.Error(1)
}

func (m *mockClusterController) MemberRemove(ctx context.Context, id uint64) (*clientv3.MemberRemoveResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*clientv3.MemberRemoveResponse), args.Error(1)
}

func (m *mockClusterController) MemberUpdate(ctx context.Context, id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error) {
	args := m.Called(ctx, id, peerAddrs)
	return args.Get(0).(*clientv3.MemberUpdateResponse), args.Error(1)
}

func (m *mockClusterController) ClusterID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}

func newClusterTest(t *testing.T) (*mockClusterController, *httptest.Server) {
	controller := &mockClusterController{}
	clusterRouter := NewClusterRouter(controller)
	router := mux.NewRouter()
	clusterRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestClusterRouterMemberList(t *testing.T) {
	ctrl, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)
	ctrl.On("MemberList", mock.Anything).Return(new(clientv3.MemberListResponse), nil)

	endpoint := "/cluster/members"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	ctrl.AssertCalled(t, "MemberList", mock.Anything)
}

func TestClusterRouterMemberAdd(t *testing.T) {
	ctrl, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)
	ctrl.On("MemberAdd", mock.Anything, mock.Anything).Return(new(clientv3.MemberAddResponse), nil)

	endpoint := "/cluster/members?peer-addrs=127.0.0.1:1234"
	req := newRequest(t, http.MethodPost, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	ctrl.AssertCalled(t, "MemberAdd", mock.Anything, []string{"127.0.0.1:1234"})
}

func TestClusterRouterMemberAddBadRequest(t *testing.T) {
	_, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	endpoint := "/cluster/members?peer-urls=127.0.0.1:1234"
	req := newRequest(t, http.MethodPost, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status (want 400): %d (%q)", resp.StatusCode, string(body))
	}
}

func TestClusterRouterMemberRemove(t *testing.T) {
	ctrl, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)
	ctrl.On("MemberRemove", mock.Anything, uint64(1234)).Return(new(clientv3.MemberRemoveResponse), nil)

	endpoint := fmt.Sprintf("/cluster/members/%x", 1234)
	req := newRequest(t, http.MethodDelete, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	ctrl.AssertCalled(t, "MemberRemove", mock.Anything, uint64(1234))
}

func TestClusterRouterMemberRemoveBadRequestNotANumber(t *testing.T) {
	_, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	endpoint := "/cluster/members/not-a-number"
	req := newRequest(t, http.MethodDelete, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status (want 400): %d (%q)", resp.StatusCode, string(body))
	}
}

func TestClusterRouterMemberRemoveBadRequestNegativeNumber(t *testing.T) {
	_, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	endpoint := "/cluster/members/-1"
	req := newRequest(t, http.MethodDelete, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status (want 400): %d (%q)", resp.StatusCode, string(body))
	}
}

func TestClusterRouterMemberUpdate(t *testing.T) {
	ctrl, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)
	ctrl.On("MemberUpdate", mock.Anything, uint64(1234), []string{"127.0.0.1:5678"}).Return(new(clientv3.MemberUpdateResponse), nil)

	endpoint := fmt.Sprintf("/cluster/members/%x?peer-addrs=127.0.0.1:5678", 1234)
	req := newRequest(t, http.MethodPut, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	ctrl.AssertCalled(t, "MemberUpdate", mock.Anything, uint64(1234), []string{"127.0.0.1:5678"})
}

func TestClusterRouterMemberUpdateBadRequestID(t *testing.T) {
	_, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	endpoint := "/cluster/members/not-a-number"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status (want 400): %d (%q)", resp.StatusCode, string(body))
	}
}

func TestClusterRouterMemberUpdateBadRequestPeerAddrs(t *testing.T) {
	_, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	endpoint := "/cluster/members/1234?foo=bar"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status (want 400): %d (%q)", resp.StatusCode, string(body))
	}
}

func TestGetClusterID(t *testing.T) {
	controller, server := newClusterTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := uuid.New().String()
	controller.On("ClusterID", mock.Anything).Return(fixture, nil)
	endpoint := "/cluster/id"
	req := newRequest(t, http.MethodGet, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "ClusterID", mock.Anything)
}
