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
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/mock"
)

type mockCheckController struct {
	mock.Mock
}

func (m *mockCheckController) AddCheckHook(ctx context.Context, check string, hook corev2.HookList) error {
	return m.Called(ctx, check, hook).Error(0)
}

func (m *mockCheckController) RemoveCheckHook(ctx context.Context, check string, hookType string, hookName string) error {
	return m.Called(ctx, check, hookType, hookName).Error(0)
}

func (m *mockCheckController) QueueAdhocRequest(ctx context.Context, check string, req *corev2.AdhocRequest) error {
	return m.Called(ctx, check, req).Error(0)
}

func TestHttpApiChecksAdhocRequest(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	store := &mockstore.MockStore{}
	queue := &mockqueue.MockQueue{}
	adhocRequest := corev2.FixtureAdhocRequest("check1", []string{"subscription1", "subscription2"})
	checkConfig := corev2.FixtureCheckConfig("check1")
	store.On("GetCheckConfigByName", mock.Anything, "check1").Return(checkConfig, nil)
	queue.On("Enqueue", mock.Anything, mock.Anything).Return(nil)
	getter := &mockqueue.Getter{}
	getter.On("GetQueue", mock.Anything).Return(queue)
	checkController := actions.NewCheckController(store, getter)
	c := &ChecksRouter{controller: checkController}
	payload, _ := json.Marshal(adhocRequest)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(defaultCtx)
	req = mux.SetURLVars(req, map[string]string{"id": "check1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(c.adhocRequest)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned incorrect status code: %v want %v", status, http.StatusAccepted)
	}
}

func TestChecksRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	router := ChecksRouter{handlers: handlers.Handlers{
		Resource: &corev2.CheckConfig{},
		Store:    s,
	}}
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.CheckConfig{}
	fixture := corev2.FixtureCheckConfig("foo")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(fixture)...)
	tests = append(tests, listTestCases(empty)...)
	tests = append(tests, createTestCases(empty)...)
	tests = append(tests, updateTestCases(fixture)...)
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestChecksRouterCustomRoutes(t *testing.T) {
	type controllerFunc func(*mockCheckController)

	// Setup the router
	controller := &mockCheckController{}
	router := ChecksRouter{controller: controller}
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		controllerFunc controllerFunc
		wantStatusCode int
	}{
		{
			name:   "it adds a check hook to a check",
			method: http.MethodPut,
			path:   "/namespaces/default/checks/check1/hooks/non-zero",
			body:   marshal(corev2.FixtureHookList("hook1")),
			controllerFunc: func(c *mockCheckController) {
				c.On("AddCheckHook", mock.Anything, "check1", mock.AnythingOfType("v2.HookList")).Return(nil)
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "it deletes a check hook from a check",
			method: http.MethodDelete,
			path:   "/namespaces/default/checks/check1/hooks/non-zero/hook/hook1",
			controllerFunc: func(c *mockCheckController) {
				c.On("RemoveCheckHook", mock.Anything, "check1", "non-zero", "hook1").Return(nil)
			},
			wantStatusCode: http.StatusNoContent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only start the HTTP server here to prevent data races in tests
			server := httptest.NewServer(parentRouter)
			defer server.Close()

			if tt.controllerFunc != nil {
				tt.controllerFunc(controller)
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
			defer res.Body.Close()

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
