package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewSilencedController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewSilencedController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
}

func TestSilencedQuery(t *testing.T) {
	defaultCtx := context.Background()

	testCases := []struct {
		name         string
		ctx          context.Context
		params       QueryParams
		storeErr     error
		storeRecords []*types.Silenced
		expectedLen  int
		expectedErr  error
	}{
		{
			name:         "No Silenced Entries",
			ctx:          defaultCtx,
			storeRecords: []*types.Silenced{},
			expectedLen:  0,
			storeErr:     nil,
			expectedErr:  nil,
		},
		{
			name: "With Silenced Entries",
			ctx:  defaultCtx,
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:   "Filter By Subscription",
			ctx:    defaultCtx,
			params: QueryParams{"subscription": "test"},
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:   "Filter By Check",
			ctx:    defaultCtx,
			params: QueryParams{"check": "test"},
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
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
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetSilencedEntriesBySubscription", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilencedEntriesByCheckName", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilencedEntries", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()

			// Exec Query
			results, err := actions.List(tc.ctx, tc.params["subscription"], tc.params["check"])

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}
