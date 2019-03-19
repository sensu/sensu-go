package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/tessen"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTessenController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewTessenController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestCreateOrUpdateTessenConfig(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *tessen.Config
		storeErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:     "Create or update",
			ctx:      context.Background(),
			argument: tessen.DefaultConfig(),
		},
		{
			name:            "Invalid input",
			ctx:             context.Background(),
			argument:        tessen.DefaultConfig(),
			storeErr:        &store.ErrNotValid{},
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			argument:        tessen.DefaultConfig(),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewTessenController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("CreateOrUpdateTessenConfig", mock.Anything, mock.Anything).
				Return(tc.storeErr)

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
		expectedResult  *tessen.Config
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:           "Get",
			ctx:            context.Background(),
			expectedResult: tessen.DefaultConfig(),
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
		actions := NewTessenController(store)

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
