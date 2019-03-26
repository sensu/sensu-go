package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewNamespacesController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewNamespacesController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
}

func TestNamespacesQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name                  string
		ctx                   context.Context
		records               []*types.Namespace
		storeErr              error
		continueToken         string
		expectedLen           int
		expectedContinueToken string
		expectedErr           error
	}{
		{
			name: "With one org",
			ctx:  defaultCtx,
			records: []*types.Namespace{
				types.FixtureNamespace("default"),
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "Store Failure",
			ctx:         defaultCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
		{
			name: "no continue token",
			ctx:  defaultCtx,
			records: []*types.Namespace{
				types.FixtureNamespace("namespace1"),
				types.FixtureNamespace("namespace2"),
			},
			continueToken:         "",
			expectedLen:           2,
			expectedContinueToken: "",
		},
		{
			name: "base64url encode continue token",
			ctx:  defaultCtx,
			records: []*types.Namespace{
				types.FixtureNamespace("namespace1"),
				types.FixtureNamespace("namespace2"),
			},
			continueToken:         "Albert Camus",
			expectedLen:           2,
			expectedContinueToken: "QWxiZXJ0IENhbXVz",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewNamespacesController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			// Mock store methods
			store.On("ListNamespaces", tc.ctx, mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
				Return(tc.records, tc.continueToken, tc.storeErr)

			// Exec Query
			results, continueToken, err := actions.Query(tc.ctx)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.EqualValues(tc.expectedContinueToken, continueToken)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestNamespacesFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.Namespace
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No name given",
			ctx:             defaultCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureNamespace("org1"),
			argument:        "org1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "org1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewNamespacesController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetNamespace", tc.ctx, mock.Anything, mock.Anything).
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

func TestNamespacesCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badOrg := types.FixtureNamespace("org1")
	badOrg.Name = "I like turtles"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Namespace
		fetchResult     *types.Namespace
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureNamespace("org1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureNamespace("org1"),
			fetchResult: types.FixtureNamespace("org1"),
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureNamespace("org1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badOrg,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewNamespacesController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetNamespace", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateNamespace", mock.Anything, mock.Anything).
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

func TestNamespacesCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badOrg := types.FixtureNamespace("org1")
	badOrg.Name = "I like turtles"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Namespace
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureNamespace("org1"),
			expectedErr: false,
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureNamespace("org1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badOrg,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewNamespacesController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("CreateNamespace", mock.Anything, mock.Anything).
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

func TestNamespacesDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.Namespace
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			argument:    "org1",
			fetchResult: types.FixtureNamespace("org1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     types.FixtureNamespace("org1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        "org1",
			fetchResult:     types.FixtureNamespace("org1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewNamespacesController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetNamespace", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteNamespace", mock.Anything, "org1").
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
