package actions

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTessenController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	bus := &mockbus.MockBus{}
	actions := NewTessenController(store, bus)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
	assert.Equal(bus, actions.bus)
}

func TestCreateOrUpdateTessenConfig(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *corev2.TessenConfig
		storeErr        error
		busErr          error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create or update",
			ctx:      context.Background(),
			argument: corev2.DefaultTessenConfig(),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        corev2.DefaultTessenConfig(),
			storeErr:        &store.ErrNotValid{Err: errors.New("error")},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        corev2.DefaultTessenConfig(),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Bus Error",
			ctx:             context.Background(),
			argument:        corev2.DefaultTessenConfig(),
			busErr:          errors.New("the bus has a flat tire"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		actions := NewTessenController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateOrUpdateTessenConfig", mock.Anything, mock.Anything).
				Return(tc.storeErr)

			bus.On("Publish", mock.Anything, mock.Anything).Return(tc.busErr)

			err := actions.CreateOrUpdate(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestGetTessenConfig(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		storeErr        error
		expectedResult  *corev2.TessenConfig
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:           "Get",
			ctx:            context.Background(),
			expectedResult: corev2.DefaultTessenConfig(),
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		bus := &mockbus.MockBus{}
		actions := NewTessenController(store, bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("GetTessenConfig", mock.Anything, mock.Anything).
				Return(tc.expectedResult, tc.storeErr)

			result, err := actions.Get(tc.ctx)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedResult, result)
			}
		})
	}
}

func TestNewTessenMetricController(t *testing.T) {
	assert := assert.New(t)

	bus := &mockbus.MockBus{}
	actions := NewTessenMetricController(bus)

	assert.NotNil(actions)
	assert.Equal(bus, actions.bus)
}

func TestPublishTessenMetrics(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        []corev2.MetricPoint
		busErr          error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name: "Publish",
			ctx:  context.Background(),
			argument: []corev2.MetricPoint{
				corev2.MetricPoint{
					Name:  "metric",
					Value: 1,
				},
			},
		},
		{
			name: "Bus Error",
			ctx:  context.Background(),
			argument: []corev2.MetricPoint{
				corev2.MetricPoint{
					Name:  "metric",
					Value: 1,
				},
			},
			busErr:          errors.New("the bus has a flat tire"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		bus := &mockbus.MockBus{}
		actions := NewTessenMetricController(bus)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			bus.On("Publish", mock.Anything, mock.Anything).Return(tc.busErr)

			err := actions.Publish(tc.ctx, tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}
