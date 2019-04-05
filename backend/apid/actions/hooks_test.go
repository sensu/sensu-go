package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewHookController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewHookController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestHookList(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.HookConfig
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No Hooks",
			ctx:         defaultCtx,
			records:     []*types.HookConfig{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Hooks",
			ctx:  defaultCtx,
			records: []*types.HookConfig{
				types.FixtureHookConfig("hook1"),
				types.FixtureHookConfig("hook2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store Failure",
			ctx:         defaultCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		actions := NewHookController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetHookConfigs", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestHookFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.HookConfig
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No argument given",
			ctx:             defaultCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureHookConfig("hook1"),
			argument:        "hook1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "missing",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHookController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetHookConfigByName", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.record, nil)

			// Exec Query
			result, err := actions.Find(tc.ctx, tc.argument)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Find() to return a record")
		})
	}
}

func TestHookCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badHook := types.FixtureHookConfig("hook1")
	badHook.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.HookConfig
		fetchResult     *types.HookConfig
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureHookConfig("hook1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureHookConfig("hook1"),
			fetchResult: types.FixtureHookConfig("hook1"),
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureHookConfig("hook1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badHook,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHookController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetHookConfigByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateHookConfig", mock.Anything, mock.Anything).
				Return(tc.createErr)

			// Exec Query
			err := actions.CreateOrReplace(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestHookCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badHook := types.FixtureHookConfig("hook1")
	badHook.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.HookConfig
		fetchResult     *types.HookConfig
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureHookConfig("hook1"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureHookConfig("hook1"),
			fetchResult:     types.FixtureHookConfig("hook1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureHookConfig("hook1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureHookConfig("hook1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badHook,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHookController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetHookConfigByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateHookConfig", mock.Anything, mock.Anything).
				Return(tc.createErr)

			// Exec Query
			err := actions.Create(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestHookDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.HookConfig
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			argument:    "hook1",
			fetchResult: types.FixtureHookConfig("hook1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        "hook1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store Err on Delete",
			ctx:             defaultCtx,
			argument:        "hook1",
			fetchResult:     types.FixtureHookConfig("hook1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			argument:        "hook1",
			fetchResult:     types.FixtureHookConfig("hook1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHookController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetHookConfigByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteHookConfigByName", mock.Anything, tc.argument).
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}
