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

func TestNewUserController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewUserController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestUserList(t *testing.T) {
	ctxWithAuthorizedViewer := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.User
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No Users",
			ctx:         ctxWithAuthorizedViewer,
			storeErr:    nil,
			expectedLen: 0,
			expectedErr: nil,
		},
		{
			name: "With Users",
			ctx:  ctxWithAuthorizedViewer,
			records: []*types.User{
				types.FixtureUser("user1"),
				types.FixtureUser("user2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store Failure",
			ctx:         ctxWithAuthorizedViewer,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		actions := NewUserController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetAllUsers", pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestUserGet(t *testing.T) {
	ctxWithAuthorizedViewer := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		storedRecord    *types.User
		storeErr        error
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No name given",
			ctx:             ctxWithAuthorizedViewer,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:         "Found",
			ctx:          ctxWithAuthorizedViewer,
			storedRecord: types.FixtureUser("user1"),
			argument:     "user1",
			expected:     true,
		},
		{
			name:            "store Err",
			ctx:             ctxWithAuthorizedViewer,
			argument:        "user1",
			storeErr:        errors.New("test"),
			expected:        false,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Not Found",
			ctx:             ctxWithAuthorizedViewer,
			argument:        "user1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetUser", mock.Anything, tc.argument).
				Return(tc.storedRecord, tc.storeErr)

			// Exec Query
			result, err := actions.Get(tc.ctx, tc.argument)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Get() to return a record")
		})
	}
}

func TestUserCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badUser := types.FixtureUser("user1")
	badUser.Username = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.User
		fetchResult     *types.User
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureUser("user1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureUser("user1"),
			fetchResult: types.FixtureUser("user1"),
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badUser,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("UpdateUser", mock.Anything).Return(tc.createErr)
			store.
				On("GetUser", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			// Exec Query
			err := actions.CreateOrReplace(tc.ctx, tc.argument)

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
func TestUserCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badUser := types.FixtureUser("user1")
	badUser.Username = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.User
		fetchResult     *types.User
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureUser("user1"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			fetchResult:     types.FixtureUser("user1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badUser,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("UpdateUser", mock.Anything).Return(tc.createErr)
			store.
				On("GetUser", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			// Exec Query
			err := actions.Create(tc.ctx, tc.argument)

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

func TestUserDisable(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.User
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Disabled",
			ctx:         defaultCtx,
			argument:    "user1",
			fetchResult: types.FixtureUser("user1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        "user1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store Err on Update",
			ctx:             defaultCtx,
			argument:        "user1",
			fetchResult:     types.FixtureUser("user1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("UpdateUser", tc.fetchResult).
				Return(tc.updateErr)
			store.
				On("GetUser", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			// Exec Query
			err := actions.Disable(tc.ctx, tc.argument)

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

func TestUserEnable(t *testing.T) {
	correctPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	disabledUser := func() *types.User {
		user := types.FixtureUser("user1")
		user.Disabled = true
		return user
	}

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.User
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Enable",
			ctx:         correctPermsCtx,
			argument:    "user1",
			fetchResult: disabledUser(),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             correctPermsCtx,
			argument:        "user1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store Err on Update",
			ctx:             correctPermsCtx,
			argument:        "user1",
			fetchResult:     disabledUser(),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			// Mock store methods
			store.On("UpdateUser", mock.Anything).Return(tc.updateErr).Once()
			store.
				On("GetUser", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			// Exec Query
			err := actions.Enable(tc.ctx, tc.argument)

			// Assertions
			assert := assert.New(t)
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
