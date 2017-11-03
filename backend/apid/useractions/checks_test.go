package useractions

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

func TestNewCheckActions(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewCheckActions(nil, store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
	assert.Nil(actions.Context)

	ctx := context.TODO()
	actions = actions.WithContext(ctx)

	assert.Equal(ctx, actions.Context)
}

func TestCheckQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCheck, types.RulePermRead),
		),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.CheckConfig
		params      QueryParams
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Checks",
			ctx:         defaultCtx,
			records:     []*types.CheckConfig{},
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Checks",
			ctx:  defaultCtx,
			records: []*types.CheckConfig{
				types.FixtureCheckConfig("check1"),
				types.FixtureCheckConfig("check2"),
			},
			params:      QueryParams{},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeCheck, types.RulePermCreate),
			)),
			records: []*types.CheckConfig{
				types.FixtureCheckConfig("check1"),
				types.FixtureCheckConfig("check2"),
			},
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: NewErrorf(PermissionDenied, "cannot list resources"),
		},
		{
			name:        "Store Failure",
			ctx:         defaultCtx,
			records:     nil,
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewCheckActions(nil, store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetCheckConfigs", tc.ctx).Return(tc.records, tc.storeErr)

			// Exec Query
			fetcher := actions.WithContext(tc.ctx)
			results, err := fetcher.Query(tc.params)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestCheckFind(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithRules(
		testutil.ContextWithOrgEnv("default", "default"),
		types.FixtureRuleWithPerms(types.RuleTypeCheck, types.RulePermRead),
	))

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.CheckConfig
		params          QueryParams
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No Params",
			ctx:             defaultCtx,
			params:          QueryParams{},
			expected:        false,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureCheckConfig("check1"),
			params:          QueryParams{"id": "check1"},
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			params:          QueryParams{"id": "check1"},
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeCheck, types.RulePermCreate),
			)),
			record:          types.FixtureCheckConfig("check1"),
			params:          QueryParams{"id": "check1"},
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewCheckActions(nil, store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetCheckConfigByName", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.record, nil)

			// Exec Query
			fetcher := actions.WithContext(tc.ctx)
			result, err := fetcher.Find(tc.params)

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
