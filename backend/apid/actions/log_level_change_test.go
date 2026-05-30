package actions

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/util/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewLogLevelChangeController(t *testing.T) {
	assert := assert.New(t)

	controller := NewLogLevelChangeController()
	assert.NotNil(controller)
}

func TestSetGlobalModuleLogLevel(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		logLevel       string
		expectedError  bool
		expectedResult bool
	}{
		{
			name:           "Valid log level",
			logLevel:       "info",
			expectedError:  false,
			expectedResult: true,
		},
		{
			name:           "Empty log level",
			logLevel:       "",
			expectedError:  true,
			expectedResult: false,
		},
		{
			name:           "Debug log level",
			logLevel:       "debug",
			expectedError:  false,
			expectedResult: true,
		},
		{
			name:           "Warn log level",
			logLevel:       "warn",
			expectedError:  false,
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			controller := NewLogLevelChangeController()
			result, err := controller.SetGlobalModuleLogLevel(ctx, tc.logLevel)

			if tc.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				if tc.expectedResult {
					assert.NotEmpty(result)
				}
			}
		})
	}
}

func TestSetGlobalModuleLogLevel_WithEtcdModule(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()
	result, err := controller.SetGlobalModuleLogLevel(ctx, "debug")

	// Should handle etcd module specially
	assert.NoError(err)
	assert.NotEmpty(result)

	// Check if etcd module is in the results
	hasEtcd := false
	for _, change := range result {
		if change.Module == "etcd" {
			hasEtcd = true
			break
		}
	}
	assert.True(hasEtcd, "Should include etcd module in results")
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name          string
		requests      []logging.LogLevelRequest
		expectedError bool
		expectedLen   int
	}{
		{
			name: "Single valid request",
			requests: []logging.LogLevelRequest{
				{Module: "apid", Level: "info"},
			},
			expectedError: false,
			expectedLen:   1,
		},
		{
			name:          "Empty requests",
			requests:      []logging.LogLevelRequest{},
			expectedError: false,
			expectedLen:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			controller := NewLogLevelChangeController()
			result, err := controller.Create(ctx, tc.requests)

			assert.NoError(err)
			assert.Len(result, tc.expectedLen)

			// Check that each result has the expected structure
			for _, res := range result {
				assert.NotEmpty(res.Module)
				assert.NotEmpty(res.OldLevel)
				assert.NotEmpty(res.NewLevel)
			}
		})
	}
}

func TestCreate_WithEtcdModule(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()
	requests := []logging.LogLevelRequest{
		{Module: "etcd", Level: "debug"},
	}

	result, err := controller.Create(ctx, requests)

	assert.NoError(err)
	assert.Len(result, 1)
	assert.Equal("etcd", result[0].Module)
	assert.Equal("debug", result[0].NewLevel)
}

func TestCreate_WithErrorHandling(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()
	requests := []logging.LogLevelRequest{
		{Module: "apid", Level: "invalid-level"},
	}

	result, err := controller.Create(ctx, requests)

	// Should not return error, but mark individual requests as failed
	assert.NoError(err)
	assert.Len(result, 1)
	assert.NotEmpty(result[0].Error, "Should have error for invalid log level")
}

func TestList(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		expectedError  bool
		expectedResult bool
	}{
		{
			name:           "List all modules",
			expectedError:  false,
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			controller := NewLogLevelChangeController()
			result, err := controller.List(ctx)

			if tc.expectedError {
				assert.Error(err)
				assert.Empty(result)
			} else {
				assert.NoError(err)
				if tc.expectedResult {
					assert.NotEmpty(result)
				}
			}
		})
	}
}

func TestModuleLogLevel(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		module         string
		expectedError  bool
		expectedResult bool
	}{
		{
			name:           "Valid module",
			module:         "apid",
			expectedError:  false,
			expectedResult: true,
		},
		{
			name:           "Etcd module",
			module:         "etcd",
			expectedError:  false,
			expectedResult: true,
		},
		{
			name:           "Invalid module",
			module:         "invalid-module",
			expectedError:  true,
			expectedResult: false,
		},
		{
			name:           "Empty module",
			module:         "",
			expectedError:  true,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			controller := NewLogLevelChangeController()
			result, err := controller.ModuleLogLevel(ctx, tc.module)

			if tc.expectedError {
				assert.Error(err)
				assert.Nil(result)
			} else {
				assert.NoError(err)
				if tc.expectedResult {
					assert.NotNil(result)
					assert.Equal(tc.module, result.Module)
					assert.NotEmpty(result.Level)
				}
			}
		})
	}
}

func TestModuleLogLevel_NotFound(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()
	result, err := controller.ModuleLogLevel(ctx, "non-existent-module")

	assert.Error(err)
	assert.Contains(err.Error(), "no module log levels found")
	assert.Nil(result)
}

func TestLogLevelChangeController_Integration(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()

	// Test the full workflow
	// 1. List all modules
	listResult, err := controller.List(ctx)
	if err == nil {
		assert.NotEmpty(listResult)

		// 2. Get specific module
		if len(listResult) > 0 {
			module := listResult[0].Module
			moduleResult, err := controller.ModuleLogLevel(ctx, module)
			assert.NoError(err)
			assert.Equal(module, moduleResult.Module)

			// 3. Set log level for that module
			createResult, err := controller.Create(ctx, []logging.LogLevelRequest{
				{Module: module, Level: "debug"},
			})
			assert.NoError(err)
			assert.Len(createResult, 1)
			assert.Equal(module, createResult[0].Module)
		}
	}
}

func TestCreate_EdgeCases(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	testCases := []struct {
		name          string
		requests      []logging.LogLevelRequest
		expectedLen   int
		expectedError bool
	}{
		{
			name:          "Nil requests",
			requests:      nil,
			expectedLen:   0,
			expectedError: false,
		},
		{
			name: "Mixed valid and invalid requests",
			requests: []logging.LogLevelRequest{
				{Module: "apid", Level: "info"},
				{Module: "invalid", Level: "invalid-level"},
				{Module: "etcd", Level: "debug"},
			},
			expectedLen:   3,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewLogLevelChangeController()
			result, err := controller.Create(ctx, tc.requests)

			assert.NoError(err)
			assert.Len(result, tc.expectedLen)
		})
	}
}

func TestLogLevelChangeController_Concurrency(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	controller := NewLogLevelChangeController()

	// Test concurrent access to the controller
	done := make(chan bool, 2)

	go func() {
		result, err := controller.SetGlobalModuleLogLevel(ctx, "info")
		assert.NoError(err)
		assert.NotEmpty(result)
		done <- true
	}()

	go func() {
		result, err := controller.List(ctx)
		// List might fail if no modules are configured
		if err == nil {
			assert.NotEmpty(result)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}
