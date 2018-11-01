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

func TestNewRoleController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewRoleController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestRoleQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))
	badCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
	))

	testCases := []struct {
		name         string
		ctx          context.Context
		storeRecords []*types.Role
		storeErr     error
		expectedLen  int
		expectedErr  error
	}{
		{
			name: "With Roles",
			ctx:  defaultCtx,
			storeRecords: []*types.Role{
				simpleRoleFixture(),
				simpleRoleFixture(),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx:  badCtx,
			storeRecords: []*types.Role{
				simpleRoleFixture(),
				simpleRoleFixture(),
			},
			expectedLen: 0,
			storeErr:    nil,
		},
		{
			name:         "Store Failure",
			ctx:          defaultCtx,
			storeRecords: nil,
			expectedLen:  0,
			storeErr:     errors.New(""),
			expectedErr:  NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetRoles", tc.ctx).Return(tc.storeRecords, tc.storeErr)

			// Exec Query
			results, err := actions.Query(tc.ctx)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestRoleFind(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))
	badCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
	))

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		storeRecord     *types.Role
		storeErr        error
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
			storeRecord:     simpleRoleFixture(),
			argument:        "check1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			argument:        "missing",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err",
			ctx:             defaultCtx,
			argument:        "err",
			storeErr:        errors.New("etcd caught fire"),
			expected:        false,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Read Permission",
			ctx:             badCtx,
			storeRecord:     simpleRoleFixture(),
			argument:        "check1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", tc.ctx, mock.Anything).
				Return(tc.storeRecord, tc.storeErr)

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

func TestRoleCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
		types.RulePermUpdate,
	))
	badCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
	))

	badRole := simpleRoleFixture()
	badRole.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Role
		fetchResult     *types.Role
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    simpleRoleFixture(),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    simpleRoleFixture(),
			fetchResult: simpleRoleFixture(),
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             badCtx,
			argument:        simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badRole,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateRole", mock.Anything, mock.Anything).
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

func TestRoleCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
	))
	badCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))

	badRole := simpleRoleFixture()
	badRole.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Role
		fetchResult     *types.Role
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    simpleRoleFixture(),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             badCtx,
			argument:        simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badRole,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateRole", mock.Anything, mock.Anything).
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

func TestRoleUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermUpdate,
	))
	wrongPermsCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))

	badRole := simpleRoleFixture()
	badRole.Rules[0].Permissions = []string{}

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Role
		fetchResult     *types.Role
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    simpleRoleFixture(),
			fetchResult: simpleRoleFixture(),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			fetchResult:     simpleRoleFixture(),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        simpleRoleFixture(),
			fetchResult:     simpleRoleFixture(),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        simpleRoleFixture(),
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badRole,
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, tc.argument.Name).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateRole", mock.Anything, mock.Anything).
				Return(tc.updateErr)

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

func TestRoleAddRule(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermUpdate,
	))
	wrongPermsCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))

	simpleRule := func() types.Rule {
		return types.FixtureRuleWithPerms("x", "create")
	}

	testCases := []struct {
		name            string
		ctx             context.Context
		nameArg         string
		ruleArg         types.Rule
		fetchResult     *types.Role
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			nameArg:     "checks",
			ruleArg:     simpleRule(),
			fetchResult: simpleRoleFixture(),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         simpleRule(),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         simpleRule(),
			fetchResult:     simpleRoleFixture(),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         simpleRule(),
			fetchResult:     simpleRoleFixture(),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			nameArg:         "checks",
			ruleArg:         simpleRule(),
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         types.FixtureRuleWithPerms("x"), // No perms.
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, tc.nameArg).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateRole", mock.Anything, mock.Anything).
				Return(tc.updateErr)

			// Exec Query
			err := actions.AddRule(tc.ctx, tc.nameArg, tc.ruleArg)

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

func TestRoleRemoveRule(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermUpdate,
	))
	wrongPermsCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermRead,
	))

	simpleRole := func() *types.Role {
		role := simpleRoleFixture()
		role.Name = "checks"
		role.Rules = []types.Rule{
			types.FixtureRuleWithPerms("my-rule", "create"),
		}
		return role
	}

	testCases := []struct {
		name            string
		ctx             context.Context
		nameArg         string
		ruleArg         string
		fetchResult     *types.Role
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			nameArg:     "checks",
			ruleArg:     "my-rule",
			fetchResult: simpleRole(),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         "my-rule",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         "my-rule",
			fetchResult:     simpleRoleFixture(),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			nameArg:         "checks",
			ruleArg:         "my-rule",
			fetchResult:     simpleRoleFixture(),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			nameArg:         "checks",
			ruleArg:         "my-rule",
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, tc.nameArg).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateRole", mock.Anything, mock.Anything).
				Return(tc.updateErr)

			// Exec Query
			err := actions.RemoveRule(tc.ctx, tc.nameArg, tc.ruleArg)

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

func TestRoleDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermDelete,
	))
	badCtx := testutil.NewContext(testutil.ContextWithPerms(
		types.RuleTypeRole,
		types.RulePermCreate,
	))

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.Role
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			argument:    "role1",
			fetchResult: simpleRoleFixture(),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        "role1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			argument:        "role1",
			fetchResult:     simpleRoleFixture(),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        "role1",
			fetchResult:     simpleRoleFixture(),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             badCtx,
			argument:        "role1",
			fetchResult:     simpleRoleFixture(),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewRoleController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetRoleByName", mock.Anything, tc.argument).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteRoleByName", mock.Anything, mock.Anything).
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

func simpleRoleFixture() *types.Role {
	return types.FixtureRole("a", "b")
}
