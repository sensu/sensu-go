package useractions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEventActions(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	eventActions := NewEventActions(store)

	assert.NotNil(eventActions)
	assert.Equal(store, eventActions.Store)
	assert.NotNil(eventActions.Policy)
	assert.Nil(eventActions.Context)

	ctx := context.TODO()
	eventActions = eventActions.WithContext(ctx)

	assert.Equal(ctx, eventActions.Context)
}

func TestQuery(t *testing.T) {
	testCases := []struct {
		name        string
		events      []*types.Event
		params      QueryParams
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Params No Events",
			events:      []*types.Event{},
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Events",
			events: []*types.Event{
				types.FixtureEvent("entity1", "check1"),
				types.FixtureEvent("entity2", "check2"),
			},
			params:      QueryParams{},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Entity Param",
			events: []*types.Event{
				types.FixtureEvent("entity1", "check1"),
			},
			params: QueryParams{
				"entity": "entity1",
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:   "Store Failure",
			events: nil,
			params: QueryParams{
				"entity": "entity1",
			},
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.TODO()

			store := &mockstore.MockStore{}
			store.On("GetEvents", ctx).Return(tc.events, tc.storeErr)
			store.On("GetEventsByEntity", ctx, mock.Anything).Return(tc.events, tc.storeErr)

			eventActions := NewEventActions(store).WithContext(ctx)
			results, err := eventActions.Query(tc.params)

			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestFind(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(true, true)
}
