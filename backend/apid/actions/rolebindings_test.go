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

func TestNewRoleBindingController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewRoleBindingController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
}

func TestRoleBindingCreate(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.RoleBinding
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create",
			ctx:      context.Background(),
			argument: types.FixtureRoleBinding("read-write", "default"),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Already existing",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        &store.ErrAlreadyExists{},
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateRoleBinding", mock.Anything, mock.Anything).
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

func TestRoleBindingCreateOrReplace(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.RoleBinding
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create or update",
			ctx:      context.Background(),
			argument: types.FixtureRoleBinding("read-write", "default"),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateOrUpdateRoleBinding", mock.Anything, mock.Anything).
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

func TestRoleBindingDestroy(t *testing.T) {
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
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("DeleteRoleBinding", mock.Anything, mock.Anything).
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

func TestRoleBindingGet(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		storeErr        error
		expectedResult  *types.RoleBinding
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:           "Get",
			ctx:            context.Background(),
			argument:       "read-write",
			expectedResult: types.FixtureRoleBinding("read-write", "default"),
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
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("GetRoleBinding", mock.Anything, mock.Anything).
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

func TestRoleBindingList(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		storeErr        error
		expectedResult  []*types.RoleBinding
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name: "List",
			ctx:  context.Background(),
			expectedResult: []*types.RoleBinding{
				types.FixtureRoleBinding("read-write", "default"),
				types.FixtureRoleBinding("read-only", "default"),
				types.FixtureRoleBinding("sysadmin", "IT"),
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
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("ListRoleBindings", mock.Anything).
				Return(tc.expectedResult, tc.storeErr)

			result, err := actions.List(tc.ctx)

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

func TestRoleBindingUpdate(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.RoleBinding
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Update",
			ctx:      context.Background(),
			argument: types.FixtureRoleBinding("read-write", "default"),
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        types.FixtureRoleBinding("read-write", "default"),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleBindingController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("UpdateRoleBinding", mock.Anything, mock.Anything).
				Return(tc.storeErr)

			err := actions.Update(tc.ctx, *tc.argument)

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
