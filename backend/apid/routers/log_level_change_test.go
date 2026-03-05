package routers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLogLevelChangeController struct {
	mock.Mock
}

func (m *mockLogLevelChangeController) Create(ctx context.Context, requests []logging.LogLevelRequest) ([]logging.LogHistory, error) {
	args := m.Called(ctx, requests)
	return args.Get(0).([]logging.LogHistory), args.Error(1)
}

func (m *mockLogLevelChangeController) List(ctx context.Context) ([]logging.LogLevelRequest, error) {
	args := m.Called(ctx)
	return args.Get(0).([]logging.LogLevelRequest), args.Error(1)
}

func (m *mockLogLevelChangeController) ModuleLogLevel(ctx context.Context, module string) (*logging.LogLevelRequest, error) {
	args := m.Called(ctx, module)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logging.LogLevelRequest), args.Error(1)
}

func (m *mockLogLevelChangeController) SetGlobalModuleLogLevel(ctx context.Context, logLevel string) ([]*logging.LogHistory, error) {
	args := m.Called(ctx, logLevel)
	return args.Get(0).([]*logging.LogHistory), args.Error(1)
}

func TestLogLevelChangeRouter(t *testing.T) {
	type controllerFunc func(*mockLogLevelChangeController)

	// Setup the router
	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
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
			name:   "POST /loglevel - create log level changes",
			method: http.MethodPost,
			path:   "/api/core/v2/loglevel",
			body:   []byte(`[{"module":"apid","level":"debug"}]`),
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedRequests := []logging.LogLevelRequest{
					{Module: "apid", Level: "debug"},
				}
				expectedResults := []logging.LogHistory{
					{Module: "apid", OldLevel: "info", NewLevel: "debug"},
				}
				c.On("Create", mock.Anything, expectedRequests).
					Return(expectedResults, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "GET /loglevel/modules - list all modules",
			method: http.MethodGet,
			path:   "/api/core/v2/loglevel/modules",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResults := []logging.LogLevelRequest{
					{Module: "apid", Level: "info"},
					{Module: "etcd", Level: "debug"},
				}
				c.On("List", mock.Anything).
					Return(expectedResults, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "GET /loglevel/modules/apid - get specific module log level",
			method: http.MethodGet,
			path:   "/api/core/v2/loglevel/modules/apid",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResult := &logging.LogLevelRequest{
					Module: "apid",
					Level:  "info",
				}
				c.On("ModuleLogLevel", mock.Anything, "apid").
					Return(expectedResult, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "POST /loglevel/modules/debug - set global log level",
			method: http.MethodPost,
			path:   "/api/core/v2/loglevel/modules/debug",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResults := []*logging.LogHistory{
					{Module: "apid", OldLevel: "info", NewLevel: "debug"},
					{Module: "etcd", OldLevel: "warn", NewLevel: "debug"},
				}
				c.On("SetGlobalModuleLogLevel", mock.Anything, "debug").
					Return(expectedResults, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve the request
			parentRouter.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(tt.wantStatusCode, w.Code)

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_Create(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}

	tests := []struct {
		name           string
		body           []byte
		controllerFunc func(*mockLogLevelChangeController)
		expectedError  bool
	}{
		{
			name: "Valid request",
			body: []byte(`[{"module":"apid","level":"debug"}]`),
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedRequests := []logging.LogLevelRequest{
					{Module: "apid", Level: "debug"},
				}
				expectedResults := []logging.LogHistory{
					{Module: "apid", OldLevel: "info", NewLevel: "debug"},
				}
				c.On("Create", mock.Anything, expectedRequests).
					Return(expectedResults, nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "Multiple valid requests",
			body: []byte(`[{"module":"apid","level":"debug"},{"module":"etcd","level":"info"}]`),
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedRequests := []logging.LogLevelRequest{
					{Module: "apid", Level: "debug"},
					{Module: "etcd", Level: "info"},
				}
				expectedResults := []logging.LogHistory{
					{Module: "apid", OldLevel: "info", NewLevel: "debug"},
					{Module: "etcd", OldLevel: "warn", NewLevel: "info"},
				}
				c.On("Create", mock.Anything, expectedRequests).
					Return(expectedResults, nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "Invalid JSON body",
			body: []byte(`invalid json`),
			controllerFunc: func(c *mockLogLevelChangeController) {
				// No expectations - should fail before reaching controller
			},
			expectedError: true,
		},
		{
			name: "Controller error",
			body: []byte(`[{"module":"apid","level":"debug"}]`),
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedRequests := []logging.LogLevelRequest{
					{Module: "apid", Level: "debug"},
				}
				c.On("Create", mock.Anything, expectedRequests).
					Return([]logging.LogHistory{}, actions.NewError(actions.InternalErr, errors.New("internal error"))).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/core/v2/loglevel", bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Call the handler
			result, err := router.create(req)

			if tt.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				assert.NotNil(result)
			}

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_List(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}

	tests := []struct {
		name           string
		controllerFunc func(*mockLogLevelChangeController)
		expectedError  bool
	}{
		{
			name: "Successful list",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResults := []logging.LogLevelRequest{
					{Module: "apid", Level: "info"},
					{Module: "etcd", Level: "debug"},
				}
				c.On("List", mock.Anything).
					Return(expectedResults, nil).
					Once()
			},
			expectedError: false,
		},
		{
			name: "Controller error",
			controllerFunc: func(c *mockLogLevelChangeController) {
				c.On("List", mock.Anything).
					Return([]logging.LogLevelRequest{}, actions.NewError(actions.InternalErr, errors.New("internal error"))).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/core/v2/loglevel/modules", nil)

			// Call the handler
			result, err := router.list(req)

			if tt.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				assert.NotNil(result)
			}

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_ModuleLogLevel(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}

	tests := []struct {
		name           string
		module         string
		controllerFunc func(*mockLogLevelChangeController)
		expectedError  bool
	}{
		{
			name:   "Valid module",
			module: "apid",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResult := &logging.LogLevelRequest{
					Module: "apid",
					Level:  "info",
				}
				c.On("ModuleLogLevel", mock.Anything, "apid").
					Return(expectedResult, nil).
					Once()
			},
			expectedError: false,
		},
		{
			name:   "Module not found",
			module: "nonexistent",
			controllerFunc: func(c *mockLogLevelChangeController) {
				c.On("ModuleLogLevel", mock.Anything, "nonexistent").
					Return(nil, actions.NewError(actions.NotFound, errors.New("not found"))).
					Once()
			},
			expectedError: true,
		},
		{
			name:   "Controller error",
			module: "apid",
			controllerFunc: func(c *mockLogLevelChangeController) {
				c.On("ModuleLogLevel", mock.Anything, "apid").
					Return(nil, actions.NewError(actions.InternalErr, errors.New("internal error"))).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request with URL parameters
			req := httptest.NewRequest(http.MethodGet, "/api/core/v2/loglevel/modules/"+tt.module, nil)
			req = mux.SetURLVars(req, map[string]string{"module": tt.module})

			// Call the handler
			result, err := router.moduleLogLevel(req)

			if tt.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				assert.NotNil(result)
			}

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_SetGlobalModuleLogLevel(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}

	tests := []struct {
		name           string
		logLevel       string
		controllerFunc func(*mockLogLevelChangeController)
		expectedError  bool
	}{
		{
			name:     "Valid log level",
			logLevel: "debug",
			controllerFunc: func(c *mockLogLevelChangeController) {
				expectedResults := []*logging.LogHistory{
					{Module: "apid", OldLevel: "info", NewLevel: "debug"},
					{Module: "etcd", OldLevel: "warn", NewLevel: "debug"},
				}
				c.On("SetGlobalModuleLogLevel", mock.Anything, "debug").
					Return(expectedResults, nil).
					Once()
			},
			expectedError: false,
		},
		{
			name:     "Controller error",
			logLevel: "invalid",
			controllerFunc: func(c *mockLogLevelChangeController) {
				c.On("SetGlobalModuleLogLevel", mock.Anything, "invalid").
					Return([]*logging.LogHistory{}, actions.NewError(actions.InvalidArgument, errors.New("invalid argument"))).
					Once()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request with URL parameters
			req := httptest.NewRequest(http.MethodPost, "/api/core/v2/loglevel/modules/"+tt.logLevel, nil)
			req = mux.SetURLVars(req, map[string]string{"loglevel": tt.logLevel})

			// Call the handler
			result, err := router.SetGlobalModuleLogLevel(req)

			if tt.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				assert.NotNil(result)
			}

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_ErrorHandling(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		controllerFunc func(*mockLogLevelChangeController)
		expectedError  bool
	}{
		{
			name:   "Empty request body",
			method: http.MethodPost,
			path:   "/api/core/v2/loglevel",
			body:   []byte{},
			controllerFunc: func(c *mockLogLevelChangeController) {
				// No expectations - should fail before reaching controller
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller expectations
			tt.controllerFunc(controller)

			// Create request
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}

			// Set URL variables if needed
			if tt.path == "/api/core/v2/loglevel/modules/invalid%20module" {
				req = mux.SetURLVars(req, map[string]string{"module": "invalid module"})
			} else if tt.path == "/api/core/v2/loglevel/modules/invalid%20level" {
				req = mux.SetURLVars(req, map[string]string{"loglevel": "invalid level"})
			}

			// Call the appropriate handler
			var result interface{}
			var err error

			switch tt.method {
			case http.MethodGet:
				if tt.path == "/api/core/v2/loglevel/modules/invalid%20module" {
					result, err = router.moduleLogLevel(req)
				} else {
					result, err = router.list(req)
				}
			case http.MethodPost:
				if tt.path == "/api/core/v2/loglevel/modules/invalid%20level" {
					result, err = router.SetGlobalModuleLogLevel(req)
				} else {
					result, err = router.create(req)
				}
			}

			if tt.expectedError {
				assert.Error(err)
				assert.Empty(result)
			}

			// Verify controller expectations
			controller.AssertExpectations(t)
		})
	}
}

func TestLogLevelChangeRouter_NewLogLevelChangeRouter(t *testing.T) {
	assert := assert.New(t)

	router := NewLogLevelChangeRouter()

	assert.NotNil(router)
	assert.NotNil(router.controller)
}

func TestLogLevelChangeRouter_Mount(t *testing.T) {
	assert := assert.New(t)

	controller := &mockLogLevelChangeController{}
	router := LogLevelChangeRouter{controller: controller}
	parentRouter := mux.NewRouter()

	// Mount the router
	router.Mount(parentRouter)

	// Verify that routes are registered
	// This is a basic check - in a real scenario you might want to verify specific routes
	assert.NotNil(parentRouter)
}
