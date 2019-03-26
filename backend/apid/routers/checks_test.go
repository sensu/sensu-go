package routers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/store"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

type mockCheckController struct {
	mock.Mock
}

func (m *mockCheckController) Create(ctx context.Context, check types.CheckConfig) error {
	return m.Called(ctx, check).Error(0)
}

func (m *mockCheckController) CreateOrReplace(ctx context.Context, check types.CheckConfig) error {
	return m.Called(ctx, check).Error(0)
}

func (m *mockCheckController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func (m *mockCheckController) Find(ctx context.Context, check string) (*types.CheckConfig, error) {
	args := m.Called(ctx, check)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

func (m *mockCheckController) Destroy(ctx context.Context, check string) error {
	return m.Called(ctx, check).Error(0)
}

func (m *mockCheckController) AddCheckHook(ctx context.Context, check string, hook types.HookList) error {
	return m.Called(ctx, check, hook).Error(0)
}

func (m *mockCheckController) RemoveCheckHook(ctx context.Context, check string, hookType string, hookName string) error {
	return m.Called(ctx, check, hookType, hookName).Error(0)
}

func (m *mockCheckController) QueueAdhocRequest(ctx context.Context, check string, req *types.AdhocRequest) error {
	return m.Called(ctx, check, req).Error(0)
}

func newCheckTest(t *testing.T) (*mockCheckController, *httptest.Server) {
	controller := &mockCheckController{}
	checkRouter := NewChecksRouter(controller)
	router := mux.NewRouter()
	checkRouter.Mount(router)

	return controller, httptest.NewServer(router)
}

func TestPostCheck(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	check := types.FixtureCheckConfig("check1")
	controller.On("Create", mock.Anything, mock.Anything).Return(nil)
	b, _ := json.Marshal(check)
	body := bytes.NewReader(b)
	endpoint := "/namespaces/default/checks"
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
		mock.AnythingOfType("CheckConfig"))
}

func TestPutCheck(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("CreateOrReplace", mock.Anything, mock.Anything).Return(nil)
	b, _ := json.Marshal(types.FixtureCheckConfig("check1"))
	body := bytes.NewReader(b)
	endpoint := "/namespaces/default/checks/check1"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "CreateOrReplace", mock.Anything, mock.Anything)
}

func TestGetCheck(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := types.FixtureCheckConfig("check1")
	controller.On("Find", mock.Anything, "check1").Return(fixture, nil)
	endpoint := "/namespaces/default/checks/check1"
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

func TestDeleteCheck(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("Destroy", mock.Anything, "check1").Return(nil)
	endpoint := "/namespaces/default/checks/check1"
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

func TestListChecks(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	fixtures := []corev2.Resource{types.FixtureCheckConfig("check1")}
	controller.On("List", mock.Anything, &store.SelectionPredicate{}).Return(fixtures, nil)
	endpoint := "/namespaces/default/checks"
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

func TestPutCheckHook(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	fixture := types.FixtureHookList("hook1")
	controller.On("AddCheckHook", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	b, _ := json.Marshal(fixture)
	body := bytes.NewReader(b)
	endpoint := "/namespaces/default/checks/check1/hooks/non-zero"
	req := newRequest(t, http.MethodPut, server.URL+endpoint, body)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "AddCheckHook", mock.Anything, mock.Anything, mock.Anything)
}

func TestDeleteCheckHook(t *testing.T) {
	controller, server := newCheckTest(t)
	defer server.Close()

	client := new(http.Client)

	controller.On("RemoveCheckHook", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	endpoint := "/namespaces/default/checks/check1/hooks/non-zero/hook/hook1"
	req := newRequest(t, http.MethodDelete, server.URL+endpoint, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("bad status: %d (%q)", resp.StatusCode, string(body))
	}

	controller.AssertCalled(t, "RemoveCheckHook", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestHttpApiChecksAdhocRequest(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	store := &mockstore.MockStore{}
	queue := &mockqueue.MockQueue{}
	adhocRequest := types.FixtureAdhocRequest("check1", []string{"subscription1", "subscription2"})
	checkConfig := types.FixtureCheckConfig("check1")
	store.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(checkConfig, nil)
	queue.On("Enqueue", mock.Anything, mock.Anything).Return(nil)
	getter := &mockqueue.Getter{}
	getter.On("GetQueue", mock.Anything).Return(queue)
	checkController := actions.NewCheckController(store, getter)
	c := &ChecksRouter{controller: checkController}
	payload, _ := json.Marshal(adhocRequest)
	req, err := http.NewRequest(http.MethodPost, "/checks/check1/execute", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(c.adhocRequest)
	handler.ServeHTTP(rr, req.WithContext(defaultCtx))

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned incorrect status code: %v want %v", status, http.StatusAccepted)
	}
}
