package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewOrganizationsController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewOrganizationsController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestOrganizationsQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermRead),
		),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Organization
		params      QueryParams
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name: "With one org",
			ctx:  defaultCtx,
			records: []*types.Organization{
				types.FixtureOrganization("default"),
			},
			params:      QueryParams{},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeOrganization, types.RulePermCreate),
			)),
			records: []*types.Organization{
				types.FixtureOrganization("org1"),
				types.FixtureOrganization("org2"),
			},
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
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
		actions := NewOrganizationsController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			// Mock store methods
			store.On("GetOrganizations", tc.ctx).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.Query(tc.ctx, tc.params)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}
