package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewClusterRoleController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewClusterRoleController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
}

func TestClusterRoleCreate(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.ClusterRole
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create",
			ctx:      context.Background(),
			argument: types.FixtureClusterRole("read-write"),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        types.FixtureClusterRole("read-write"),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Already exists",
			ctx:             context.Background(),
			argument:        types.FixtureClusterRole("read-write"),
			storeErr:        &store.ErrAlreadyExists{},
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        types.FixtureClusterRole("read-write"),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateClusterRole", mock.Anything, mock.Anything).
				Return(tc.storeErr)

			err := actions.Create(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestClusterRoleCreateOrReplace(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.ClusterRole
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create or update",
			ctx:      context.Background(),
			argument: types.FixtureClusterRole("read-write"),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        types.FixtureClusterRole("read-write"),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        types.FixtureClusterRole("read-write"),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateOrUpdateClusterRole", mock.Anything, mock.Anything).
				Return(tc.storeErr)

			err := actions.CreateOrReplace(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestClusterRoleDestroy(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Delete",
			ctx:      context.Background(),
			argument: "read-write",
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			argument:        "read-write",
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        "read-write",
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("DeleteClusterRole", mock.Anything, mock.Anything).
				Return(tc.storeErr)

			err := actions.Destroy(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestClusterRoleGet(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		storeErr        error
		expectedResult  *types.ClusterRole
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:           "Get",
			ctx:            context.Background(),
			argument:       "read-write",
			expectedResult: types.FixtureClusterRole("read-write"),
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			argument:        "read-write",
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        "read-write",
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("GetClusterRole", mock.Anything, mock.Anything).
				Return(tc.expectedResult, tc.storeErr)

			result, err := actions.Get(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedResult, result)
			}
		})
	}
}

func TestClusterRoleList(t *testing.T) {
	testCases := []struct {
		name                  string
		ctx                   context.Context
		storeErr              error
		continueToken         string
		expectedResult        []*types.ClusterRole
		expectedContinueToken string
		expectedErr           bool
		expectedErrCode       ErrCode
	}{
		{
			name: "List",
			ctx:  context.Background(),
			expectedResult: []*types.ClusterRole{
				types.FixtureClusterRole("read-write"),
				types.FixtureClusterRole("read-only"),
				types.FixtureClusterRole("sysadmin"),
			},
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name: "no continue token",
			ctx:  context.Background(),
			expectedResult: []*types.ClusterRole{
				types.FixtureClusterRole("clusterRole1"),
				types.FixtureClusterRole("clusterRole2"),
			},
			continueToken:         "",
			expectedContinueToken: "",
		},
		{
			name: "base64url encode continue token",
			ctx:  context.Background(),
			expectedResult: []*types.ClusterRole{
				types.FixtureClusterRole("clusterRole1"),
				types.FixtureClusterRole("clusterRole2"),
			},
			continueToken:         "Albert Camus",
			expectedContinueToken: "QWxiZXJ0IENhbXVz",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("ListClusterRoles", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
				Return(tc.expectedResult, tc.continueToken, tc.storeErr)

			result, continueToken, err := actions.List(tc.ctx)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedResult, result)
				assert.Equal(tc.expectedContinueToken, continueToken)
			}
		})
	}
}
