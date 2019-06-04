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
	store := &mockstore.MockStore{}
	controller := &mockCheckController{}
	checkRouter := NewChecksRouter(controller, store)
	router := mux.NewRouter()
	checkRouter.Mount(router)

	return controller, httptest.NewServer(router)
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
	c := &ChecksRouter{checkController: checkController}
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

func TestChecksRouter(t *testing.T) {
	// Setup the router
	controller := &mockCheckController{}
	s := &mockstore.MockStore{}
	router := NewChecksRouter(controller, s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)

	pathPrefix := "checks"
	kind := "*v2.CheckConfig"
	check := corev2.FixtureCheckConfig("foo")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(pathPrefix, kind, check)...)
	tests = append(tests, listTestCases(pathPrefix, kind, []corev2.Resource{check})...)
	tests = append(tests, createTestCases(pathPrefix, kind)...)
	tests = append(tests, updateTestCases(pathPrefix, kind)...)
	tests = append(tests, deleteTestCases(pathPrefix, kind)...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only start the HTTP server here to prevent data races in tests
			server := httptest.NewServer(parentRouter)
			defer server.Close()

			if tt.storeFunc != nil {
				tt.storeFunc(s)
			}

			// Prepare the HTTP request
			client := new(http.Client)
			req, err := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Perform the HTTP request
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			// Inspect the response code
			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("ChecksRouter StatusCode = %v, wantStatusCode %v", res.StatusCode, tt.wantStatusCode)
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("error message: %q", string(body))
				return
			}
		})
	}
}
