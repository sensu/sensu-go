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

func TestNewEnvironmentController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.Store{}
	ctl := NewEnvironmentController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.Store)
	assert.NotNil(ctl.Policy)
}

func TestEnvironmentCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(
				types.RuleTypeEnvironment,
				types.RulePermCreate,
				types.RulePermUpdate,
			),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(
				types.RuleTypeEnvironment,
				types.RulePermCreate,
			),
		),
	)

	badMut := types.FixtureEnvironment("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Environment
		fetchResult     *types.Environment
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEnvironment("sleepy"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureEnvironment("sleepy"),
			fetchResult: types.FixtureEnvironment("sleepy"),
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEnvironment("sneezy"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badMut,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.Store{}
		ctl := NewEnvironmentController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetEnvironment", mock.Anything, mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateEnvironment", mock.Anything, mock.Anything).Return(test.createErr)

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

func TestEnvironmentCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermCreate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermRead),
		),
	)

	badMut := types.FixtureEnvironment("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Environment
		fetchResult     *types.Environment
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEnvironment("sleepy"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureEnvironment("sleepy"),
			fetchResult:     types.FixtureEnvironment("sleepy"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEnvironment("grumpy"),
			fetchErr:        errors.New("nein"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEnvironment("sneezy"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badMut,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.Store{}
		ctl := NewEnvironmentController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetEnvironment", mock.Anything, mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateEnvironment", mock.Anything, mock.Anything).Return(test.createErr)

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

func TestEnvironmentDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermDelete),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermCreate),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		envName         string
		fetchResult     *types.Environment
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			envName:     "env1",
			fetchResult: types.FixtureEnvironment("env1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			envName:         "env1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			envName:         "env1",
			fetchResult:     types.FixtureEnvironment("env1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			envName:         "env1",
			fetchResult:     types.FixtureEnvironment("env1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			envName:         "env1",
			fetchResult:     types.FixtureEnvironment("env1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.Store{}
		actions := NewEnvironmentController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEnvironment", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteEnvironment", mock.Anything, tc.fetchResult).
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, "default", tc.envName)

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

func TestEnvironmentUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermRead),
		),
	)

	badEnvironment := types.FixtureEnvironment("$$$")
	badEnvironment.Organization = "$$"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Environment
		fetchResult     *types.Environment
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureEnvironment("env1"),
			fetchResult: types.FixtureEnvironment("env1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureEnvironment("env1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        types.FixtureEnvironment("env1"),
			fetchResult:     types.FixtureEnvironment("env1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEnvironment("env1"),
			fetchResult:     types.FixtureEnvironment("env1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEnvironment("env1"),
			fetchResult:     types.FixtureEnvironment("env1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:        "Validation Error",
			ctx:         defaultCtx,
			argument:    badEnvironment,
			fetchResult: types.FixtureEnvironment("env1"),
			expectedErr: false, // only the description can be updated
		},
	}

	for _, tc := range testCases {
		store := &mockstore.Store{}
		actions := NewEnvironmentController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEnvironment", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateEnvironment", mock.Anything, mock.Anything).
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

func TestEnvironmentQuery(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermRead)))

	tests := []struct {
		name        string
		ctx         context.Context
		envs        []*types.Environment
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Environments",
			ctx:         readCtx,
			envs:        nil,
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Environments",
			ctx:  readCtx,
			envs: []*types.Environment{
				types.FixtureEnvironment("homer"),
				types.FixtureEnvironment("bart"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermCreate),
			)),
			envs: []*types.Environment{
				types.FixtureEnvironment("lisa"),
				types.FixtureEnvironment("maggie"),
			},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "Store Failure",
			ctx:         readCtx,
			envs:        nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, test := range tests {
		store := &mockstore.Store{}
		ctl := NewEnvironmentController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetEnvironments", test.ctx, "default").Return(test.envs, test.storeErr)

			results, err := ctl.Query(test.ctx, "default")

			assert.EqualValues(test.expectedErr, err)
			assert.Len(results, test.expectedLen)
		})
	}
}

func TestEnvironmentFind(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeEnvironment, types.RulePermRead),
	))

	tests := []struct {
		name            string
		ctx             context.Context
		env             *types.Environment
		envName         string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No env name",
			ctx:             readCtx,
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Found",
			ctx:             readCtx,
			env:             types.FixtureEnvironment("abe"),
			envName:         "abe",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             readCtx,
			env:             nil,
			envName:         "fox mulder",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			env:             types.FixtureEnvironment("troy maclure"),
			envName:         "troy maclure",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.Store{}
			ctl := NewEnvironmentController(store)

			// Mock store methods
			store.On("GetEnvironment", test.ctx, "default", test.envName).
				Return(test.env, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, "default", test.envName)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}
