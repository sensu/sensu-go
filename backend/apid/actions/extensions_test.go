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

func TestNewExtensionController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewExtensionController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestExtensionList(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Extension
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No Extensions",
			ctx:         defaultCtx,
			records:     []*types.Extension{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Extensions",
			ctx:  defaultCtx,
			records: []*types.Extension{
				types.FixtureExtension("extension1"),
				types.FixtureExtension("extension2"),
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
		actions := NewExtensionController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetExtensions", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestExtensionFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.Extension
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No name given",
			ctx:             defaultCtx,
			expected:        false,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureExtension("extension1"),
			argument:        "extension1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "extension1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewExtensionController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetExtension", tc.ctx, mock.Anything, mock.Anything).
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

func TestExtensionRegister(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badExtension := types.FixtureExtension("extension1")
	badExtension.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Extension
		fetchResult     *types.Extension
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureExtension("extension1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureExtension("extension1"),
			fetchResult: types.FixtureExtension("extension1"),
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureExtension("extension1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badExtension,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewExtensionController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetExtension", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("RegisterExtension", mock.Anything, mock.Anything).
				Return(tc.createErr)

			// Exec Query
			err := actions.Register(tc.ctx, *tc.argument)

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

func TestExtensionCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badExtension := types.FixtureExtension("extension1")
	badExtension.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Extension
		fetchResult     *types.Extension
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureExtension("extension1"),
			expectedErr: false,
		},
		{
			name:        "Already exists",
			ctx:         defaultCtx,
			argument:    types.FixtureExtension("extension1"),
			fetchResult: types.FixtureExtension("extension1"),
			expectedErr: false,
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureExtension("extension1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badExtension,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewExtensionController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetExtension", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("RegisterExtension", mock.Anything, mock.Anything).
				Return(tc.createErr)

			// Exec Query
			err := actions.Register(tc.ctx, *tc.argument)

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
