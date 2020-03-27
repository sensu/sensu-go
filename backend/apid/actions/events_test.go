package actions

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEventController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	bus := &mockbus.MockBus{}
	eventController := NewEventController(store, bus)

	assert.NotNil(eventController)
	assert.Equal(store, eventController.store)
	assert.Equal(bus, eventController.bus)
}

func TestEventList(t *testing.T) {
	defaultCtx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		events      []*corev2.Event
		entity      string
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No Params No Events",
			ctx:         defaultCtx,
			events:      []*corev2.Event{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Events",
			ctx:  defaultCtx,
			events: []*corev2.Event{
				corev2.FixtureEvent("entity1", "check1"),
				corev2.FixtureEvent("entity2", "check2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Entity Param",
			ctx:  defaultCtx,
			events: []*corev2.Event{
				corev2.FixtureEvent("entity1", "check1"),
			},
			entity:      "entity1",
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store Failure",
			ctx:         defaultCtx,
			events:      nil,
			entity:      "entity1",
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		eventController := NewEventController(s, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetEvents", tc.ctx, pred).Return(tc.events, tc.storeErr)

			s.On("GetEventsByEntity", tc.ctx, mock.AnythingOfType("string"), pred).
				Return(tc.events, tc.storeErr)

			// Exec Query
			results, err := eventController.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestEventFind(t *testing.T) {
	defaultCtx := context.Background()

	testCases := []struct {
		name            string
		ctx             context.Context
		event           *corev2.Event
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
			event:           corev2.FixtureEvent("entity1", "check1"),
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
			result, err := eventController.Get(tc.ctx, tc.entity, tc.check)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Get() to return an event")
		})
	}
}

func TestEventDestroy(t *testing.T) {
	defaultCtx := context.Background()

	testCases := []struct {
		name            string
		ctx             context.Context
		event           *corev2.Event
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
			event:           corev2.FixtureEvent("entity1", "check1"),
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
			err := eventController.Delete(tc.ctx, tc.entity, tc.check)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestEventCreateOrReplace(t *testing.T) {
	defaultCtx := context.Background()

	badEvent := corev2.FixtureEvent("entity1", "check1")
	badEvent.Check.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	eventNoClass := corev2.FixtureEvent("entity1", "check1")
	eventNoClass.Entity.EntityClass = ""

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *corev2.Event
		fetchResult     *corev2.Event
		fetchErr        error
		busErr          error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    corev2.FixtureEvent("entity1", "check1"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    corev2.FixtureEvent("entity1", "check1"),
			fetchResult: corev2.FixtureEvent("entity1", "check1"),
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
			argument:        corev2.FixtureEvent("entity1", "check1"),
			busErr:          errors.New("where's the wizard"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:        "Default entity class",
			ctx:         defaultCtx,
			argument:    eventNoClass,
			expectedErr: false,
		},
		{
			name:        "Keepalive",
			ctx:         defaultCtx,
			argument:    corev2.FixtureEvent("entity", "keepalive"),
			expectedErr: false,
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

			bus.On("Publish", messaging.TopicEventRaw, mock.Anything).Return(tc.busErr)
			bus.On("Publish", messaging.TopicKeepalive, mock.Anything).Return(tc.busErr)

			// Exec Query
			err := actions.CreateOrReplace(tc.ctx, tc.argument)
			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			bus.AssertCalled(t, "Publish", messaging.TopicEventRaw, tc.argument)
			if tc.argument.Check.Name == "keepalive" {
				bus.AssertCalled(t, "Publish", messaging.TopicKeepalive, tc.argument)
			}
		})
	}
}

func TestEventCreatedBy(t *testing.T) {
	claims, err := jwt.NewClaims(&corev2.User{Username: "admin"})
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), corev2.ClaimsKey, claims)
	event := corev2.FixtureEvent("entity1", "check1")

	store := &mockstore.MockStore{}
	bus := &mockbus.MockBus{}
	actions := NewEventController(store, bus)

	store.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return(event, nil)
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	err = actions.CreateOrReplace(ctx, event)
	assert.NoError(t, err)
	assert.Equal(t, "admin", event.CreatedBy)
	assert.Equal(t, "admin", event.Check.CreatedBy)
	assert.Equal(t, "admin", event.Entity.CreatedBy)
}
