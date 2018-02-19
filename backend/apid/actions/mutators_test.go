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

func TestNewMutatorController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	ctl := NewMutatorController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.Store)
	assert.NotNil(ctl.Policy)
}

func TestMutatorCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermCreate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead),
		),
	)

	badMut := types.FixtureMutator("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Mutator
		fetchResult     *types.Mutator
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureMutator("sleepy"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("sleepy"),
			fetchResult:     types.FixtureMutator("sleepy"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("grumpy"),
			fetchErr:        errors.New("nein"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureMutator("sneezy"),
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
		store := &mockstore.MockStore{}
		ctl := NewMutatorController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateMutator", mock.Anything, mock.Anything).Return(test.createErr)

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

func TestMutatorDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermDelete),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermCreate),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		mutator         string
		fetchResult     *types.Mutator
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			mutator:     "mutator1",
			fetchResult: types.FixtureMutator("mutator1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     types.FixtureMutator("mutator1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     types.FixtureMutator("mutator1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			mutator:         "mutator1",
			fetchResult:     types.FixtureMutator("mutator1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewMutatorController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteMutatorByName", mock.Anything, "mutator1").
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.mutator)

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

func TestMutatorUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead),
		),
	)

	badMutator := types.FixtureMutator("mutator1")
	badMutator.Command = ""

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Mutator
		fetchResult     *types.Mutator
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureMutator("mutator1"),
			fetchResult: types.FixtureMutator("mutator1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("mutator1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("mutator1"),
			fetchResult:     types.FixtureMutator("mutator1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("mutator1"),
			fetchResult:     types.FixtureMutator("mutator1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureMutator("mutator1"),
			fetchResult:     types.FixtureMutator("mutator1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badMutator,
			fetchResult:     types.FixtureMutator("mutator1"),
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewMutatorController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateMutator", mock.Anything, mock.Anything).
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

func TestMutatorQuery(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead)))

	tests := []struct {
		name        string
		ctx         context.Context
		mutators    []*types.Mutator
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Params, No Mutators",
			ctx:         readCtx,
			mutators:    nil,
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Mutators",
			ctx:  readCtx,
			mutators: []*types.Mutator{
				types.FixtureMutator("homer"),
				types.FixtureMutator("bart"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermCreate),
			)),
			mutators: []*types.Mutator{
				types.FixtureMutator("lisa"),
				types.FixtureMutator("maggie"),
			},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Mutator Param",
			ctx:  readCtx,
			mutators: []*types.Mutator{
				types.FixtureMutator("mr. burns"),
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "Store Failure",
			ctx:         readCtx,
			mutators:    nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewMutatorController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetMutators", test.ctx).Return(test.mutators, test.storeErr)

			results, err := ctl.Query(test.ctx)

			assert.EqualValues(test.expectedErr, err)
			assert.Len(results, test.expectedLen)
		})
	}
}

func TestMutatorFind(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead),
	))

	tests := []struct {
		name            string
		ctx             context.Context
		mutator         *types.Mutator
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "Found",
			ctx:             readCtx,
			mutator:         types.FixtureMutator("abe"),
			argument:        "abe",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             readCtx,
			mutator:         nil,
			argument:        "fox mulder",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			mutator:         types.FixtureMutator("troy maclure"),
			argument:        "troy maclure",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewMutatorController(store)

			// Mock store methods
			store.
				On("GetMutatorByName", test.ctx, test.argument).
				Return(test.mutator, nil)

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
