package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEventController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	bus := &mockbus.MockBus{}
	eventController := NewEventController(store, bus)

	assert.NotNil(eventController)
	assert.Equal(store, eventController.Store)
	assert.NotNil(eventController.Policy)
	assert.Equal(bus, eventController.Bus)
}

func TestEventQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermRead),
	))

	testCases := []struct {
		name        string
		ctx         context.Context
		events      []*types.Event
		entity      string
		check       string
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Params No Events",
			ctx:         defaultCtx,
			events:      []*types.Event{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Events",
			ctx:  defaultCtx,
			events: []*types.Event{
				types.FixtureEvent("entity1", "check1"),
				types.FixtureEvent("entity2", "check2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			events: []*types.Event{
				types.FixtureEvent("entity1", "check1"),
				types.FixtureEvent("entity2", "check2"),
			},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Entity Param",
			ctx:  defaultCtx,
			events: []*types.Event{
				types.FixtureEvent("entity1", "check1"),
			},
			entity:      "entity1",
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "Store Failure",
			ctx:         defaultCtx,
			events:      nil,
			entity:      "entity1",
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		eventController := NewEventController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetEvents", tc.ctx).Return(tc.events, tc.storeErr)
			store.On("GetEventsByEntity", tc.ctx, mock.Anything).Return(tc.events, tc.storeErr)

			// Exec Query
			results, err := eventController.Query(tc.ctx, tc.entity, tc.check)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestEventFind(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermRead),
	))

	testCases := []struct {
		name            string
		ctx             context.Context
		event           *types.Event
		entity          string
		check           string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No Params",
			ctx:             defaultCtx,
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Only Entity Param",
			ctx:             defaultCtx,
			entity:          "entity1",
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Only Check Param",
			ctx:             defaultCtx,
			check:           "check1",
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			event:           types.FixtureEvent("entity1", "check1"),
			entity:          "entity1",
			check:           "check1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			event:           nil,
			entity:          "entity1",
			check:           "check1",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			event:           types.FixtureEvent("entity1", "check1"),
			entity:          "entity1",
			check:           "check1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		eventController := NewEventController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEventByEntityCheck", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.event, nil)

			// Exec Query
			result, err := eventController.Find(tc.ctx, tc.entity, tc.check)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Find() to return an event")
		})
	}
}

func TestEventDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermDelete),
	))

	testCases := []struct {
		name            string
		ctx             context.Context
		event           *types.Event
		entity          string
		check           string
		expectedErrCode ErrCode
	}{
		{
			name:            "No Params",
			ctx:             defaultCtx,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Only Entity Param",
			ctx:             defaultCtx,
			entity:          "entity1",
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Only Check Param",
			ctx:             defaultCtx,
			check:           "check1",
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Delete",
			ctx:             defaultCtx,
			event:           types.FixtureEvent("entity1", "check1"),
			entity:          "entity1",
			check:           "check1",
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			event:           nil,
			entity:          "entity1",
			check:           "check1",
			expectedErrCode: NotFound,
		},
		{
			name: "No Delete Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			event:           types.FixtureEvent("entity1", "check1"),
			entity:          "entity1",
			check:           "check1",
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		eventController := NewEventController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEventByEntityCheck", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.event, nil)
			store.
				On("DeleteEventByEntityCheck", tc.ctx, mock.Anything, mock.Anything).
				Return(nil)

			// Exec Query
			err := eventController.Destroy(tc.ctx, tc.entity, tc.check)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestEventUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermRead),
		),
	)

	badEvent := types.FixtureEvent("entity1", "check1")
	badEvent.Check.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Event
		fetchResult     *types.Event
		fetchErr        error
		busErr          error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureEvent("entity1", "check1"),
			fetchResult: types.FixtureEvent("entity1", "check1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			fetchResult:     types.FixtureEvent("entity1", "check1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			fetchResult:     types.FixtureEvent("entity1", "check1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badEvent,
			fetchResult:     types.FixtureEvent("entity1", "check1"),
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Message Bus Error",
			ctx:             defaultCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			fetchResult:     types.FixtureEvent("entity1", "check1"),
			busErr:          errors.New("where's the wizard"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		actions := NewEventController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			bus.On("Publish", mock.Anything, mock.Anything).Return(tc.busErr)

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

func TestEventCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermRead),
		),
	)

	badEvent := types.FixtureEvent("entity1", "check1")
	badEvent.Check.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Event
		fetchResult     *types.Event
		fetchErr        error
		busErr          error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEvent("entity1", "check1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureEvent("entity1", "check1"),
			fetchResult: types.FixtureEvent("entity1", "check1"),
			expectedErr: false,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badEvent,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Message Bus Error",
			ctx:             defaultCtx,
			argument:        types.FixtureEvent("entity1", "check1"),
			busErr:          errors.New("where's the wizard"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		actions := NewEventController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)

			bus.On("Publish", mock.Anything, mock.Anything).Return(tc.busErr)

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
