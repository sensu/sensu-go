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

func TestNewHandlerController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	ctl := NewHandlerController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.store)
}

func TestHandlerCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badHandler := types.FixtureHandler("bad")
	badHandler.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Handler
		fetchResult     *types.Handler
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureHandler("foo"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureHandler("foo"),
			fetchResult:     types.FixtureHandler("foo"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureHandler("foo"),
			fetchErr:        errors.New("nein"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badHandler,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewHandlerController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetHandlerByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateHandler", mock.Anything, mock.Anything).Return(test.createErr)

			err := ctl.Create(test.ctx, *test.argument)

			if test.expectedErr {
				if cerr, ok := err.(Error); ok {
					assert.Equal(test.expectedErrCode, cerr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestHandlerCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badHandler := types.FixtureHandler("bad")
	badHandler.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	grpcHandler := types.FixtureHandler("grpc")
	grpcHandler.Type = types.HandlerGRPCType

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Handler
		fetchResult     *types.Handler
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureHandler("foo"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureHandler("foo"),
			fetchResult: types.FixtureHandler("foo"),
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badHandler,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Bad handler type",
			ctx:             defaultCtx,
			argument:        grpcHandler,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewHandlerController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetHandlerByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateHandler", mock.Anything, mock.Anything).Return(test.createErr)

			err := ctl.CreateOrReplace(test.ctx, *test.argument)

			if test.expectedErr {
				if cerr, ok := err.(Error); ok {
					assert.Equal(test.expectedErrCode, cerr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestHandlerDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.Handler
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "deleted",
			ctx:         defaultCtx,
			argument:    "handler1",
			fetchResult: types.FixtureHandler("handler1"),
			expectedErr: false,
		},
		{
			name:            "does not exist",
			ctx:             defaultCtx,
			argument:        "handler1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store err on delete",
			ctx:             defaultCtx,
			argument:        "handler1",
			fetchResult:     types.FixtureHandler("handler1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store err on fetch",
			ctx:             defaultCtx,
			argument:        "handler1",
			fetchResult:     types.FixtureHandler("handler1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHandlerController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetHandlerByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteHandlerByName", mock.Anything, "handler1").
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

func TestHandlerFind(t *testing.T) {
	readCtx := context.Background()

	tests := []struct {
		name            string
		ctx             context.Context
		handler         *types.Handler
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "no argument given",
			ctx:             readCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "found",
			ctx:             readCtx,
			handler:         types.FixtureHandler("foo"),
			argument:        "foo",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "not found",
			ctx:             readCtx,
			handler:         nil,
			argument:        "missing",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewHandlerController(store)

			// Mock store methods
			store.On("GetHandlerByName", test.ctx, test.argument).
				Return(test.handler, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, test.argument)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}

func TestHandlerList(t *testing.T) {
	readCtx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Handler
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "no params, no handlers",
			ctx:         readCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "no params with handlers",
			ctx:  readCtx,
			records: []*types.Handler{
				types.FixtureHandler("foo"),
				types.FixtureHandler("bar"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "handler name param",
			ctx:  readCtx,
			records: []*types.Handler{
				types.FixtureHandler("foo"),
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store failure",
			ctx:         readCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		actions := NewHandlerController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetHandlers", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}
