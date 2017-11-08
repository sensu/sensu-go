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

func TestMutatorQuery(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead)))

	tests := []struct {
		name        string
		ctx         context.Context
		mutators    []*types.Mutator
		params      QueryParams
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Params, No Mutators",
			ctx:         readCtx,
			mutators:    nil,
			params:      QueryParams{},
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
			params:      QueryParams{},
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
			params:      QueryParams{},
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
			params: QueryParams{
				"name": "mr. burns",
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:     "Store Failure",
			ctx:      readCtx,
			mutators: nil,
			params: QueryParams{
				"name": "ralph",
			},
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

			results, err := ctl.Query(test.ctx, test.params)

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
		params          QueryParams
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No Params",
			ctx:             readCtx,
			params:          QueryParams{},
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:    "Found",
			ctx:     readCtx,
			mutator: types.FixtureMutator("abe"),
			params: QueryParams{
				"name": "abe",
			},
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:    "Not Found",
			ctx:     readCtx,
			mutator: nil,
			params: QueryParams{
				"name": "fox mulder",
			},
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			mutator: types.FixtureMutator("troy maclure"),
			params: QueryParams{
				"name": "troy maclure",
			},
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewMutatorController(store)

			// Mock store methods
			store.On("GetMutatorByName", test.ctx, test.params["name"]).
				Return(test.mutator, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, test.params)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}
