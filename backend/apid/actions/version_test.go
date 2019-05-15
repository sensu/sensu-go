package actions

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

func TestNewVersionController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewVersionController(store, nil)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestGetVersion(t *testing.T) {
	testCases := []struct {
		name        string
		response    *corev2.Version
		expectedErr error
	}{
		{
			name:        "Good version",
			response:    corev2.FixtureVersion(),
			expectedErr: nil,
		},
		{
			name:        "Bad version",
			response:    nil,
			expectedErr: fmt.Errorf("foo"),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewVersionController(store, nil)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()

			// Mock store methods
			store.On("GetVersion", ctx, nil).Return(tc.response, tc.expectedErr)

			// Exec GetVersion
			response, err := actions.GetVersion(ctx)

			// Assert
			assert.Equal(tc.response, response)
			assert.Equal(tc.expectedErr, err)
		})
	}
}
