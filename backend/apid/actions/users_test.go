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

func TestNewUserController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewUserController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestUserQuery(t *testing.T) {
	ctxWithAuthorizedViewer := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithPerms(types.RuleTypeUser, types.RulePermRead),
	)

	testCases := []struct {
		name          string
		ctx           context.Context
		storedRecords []*types.User
		storeErr      error
		expectedLen   int
		expectedErr   error
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
			storedRecords: []*types.User{
				types.FixtureUser("user1"),
				types.FixtureUser("user2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx: testutil.NewContext(
				testutil.ContextWithPerms(types.RuleTypeUser, types.RulePermCreate),
			),
			storedRecords: []*types.User{
				types.FixtureUser("user1"),
				types.FixtureUser("user2"),
			},
			expectedLen: 0,
			storeErr:    nil,
		},
		{
			name:        "Store Failure",
			ctx:         ctxWithAuthorizedViewer,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetAllUsers").Return(tc.storedRecords, tc.storeErr)

			// Exec Query
			results, err := actions.Query(tc.ctx)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestUserFind(t *testing.T) {
	ctxWithAuthorizedViewer := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithPerms(types.RuleTypeUser, types.RulePermRead),
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
			name:            "Store Err",
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
		{
			name:            "No Read Permission",
			ctx:             testutil.NewContext(testutil.ContextWithActor("user15")),
			storedRecord:    types.FixtureUser("user1"),
			argument:        "user1",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:         "Read Themselves Without Permissions",
			ctx:          testutil.NewContext(testutil.ContextWithActor("user1")),
			storedRecord: types.FixtureUser("user1"),
			argument:     "user1",
			expected:     true,
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

func TestUserCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(
				types.RuleTypeUser,
				types.RulePermCreate,
				types.RulePermUpdate,
			),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermCreate),
		),
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
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureUser("user1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
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
func TestUserCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermCreate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermRead),
		),
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
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		// {
		// 	name:            "Store Err on Fetch",
		// 	ctx:             defaultCtx,
		// 	argument:        types.FixtureUser("user1"),
		// 	fetchErr:        errors.New("dunno"),
		// 	expectedErr:     true,
		// 	expectedErrCode: InternalErr,
		// },
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureUser("user1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
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

func TestUserUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermRead),
		),
	)

	badUser := types.FixtureUser("user1")
	badUser.Password = "1"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.User
		fetchResult     *types.User
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureUser("user1"),
			fetchResult: types.FixtureUser("user1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        types.FixtureUser("user1"),
			fetchResult:     types.FixtureUser("user1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		// {
		// 	name:            "Store Err on Fetch",
		// 	ctx:             defaultCtx,
		// 	argument:        types.FixtureUser("user1"),
		// 	fetchResult:     types.FixtureUser("user1"),
		// 	fetchErr:        errors.New("dunno"),
		// 	expectedErr:     true,
		// 	expectedErrCode: InternalErr,
		// },
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureUser("user1"),
			fetchResult:     types.FixtureUser("user1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:        "Can Set Own Password",
			ctx:         testutil.NewContext(testutil.ContextWithActor("user1")),
			argument:    &types.User{Username: "user1", Password: "12345678"},
			fetchResult: types.FixtureUser("user1"),
			expectedErr: false,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badUser,
			fetchResult:     types.FixtureUser("user1"),
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
			store.On("UpdateUser", mock.Anything).Return(tc.updateErr)
			store.
				On("GetUser", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			// Exec Query
			err := actions.Update(tc.ctx, *tc.argument)

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
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermDelete),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeUser, types.RulePermCreate),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.User
		fetchErr        error
		deleteErr       error
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
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        "user1",
			fetchResult:     types.FixtureUser("user1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		// {
		// 	name:            "Store Err on Fetch",
		// 	ctx:             defaultCtx,
		// 	argument:        "user1",
		// 	fetchResult:     types.FixtureUser("user1"),
		// 	fetchErr:        errors.New("dunno"),
		// 	expectedErr:     true,
		// 	expectedErrCode: InternalErr,
		// },
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        "user1",
			fetchResult:     types.FixtureUser("user1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewUserController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("DeleteUser", mock.Anything, tc.fetchResult).
				Return(tc.deleteErr)
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
		testutil.ContextWithPerms(types.RuleTypeUser, types.RulePermUpdate),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithPerms(types.RuleTypeUser, types.RulePermDelete),
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
			name:            "Store Err on Update",
			ctx:             correctPermsCtx,
			argument:        "user1",
			fetchResult:     disabledUser(),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		// {
		// 	name:            "Store Err on Fetch",
		// 	ctx:             correctPermsCtx,
		// 	argument:        "user1",
		// 	fetchResult:     disabledUser(),
		// 	fetchErr:        errors.New("dunno"),
		// 	expectedErr:     true,
		// 	expectedErrCode: InternalErr,
		// },
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        "user1",
			fetchResult:     disabledUser(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
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
