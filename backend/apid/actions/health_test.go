package actions

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewHealthController(store, nil, nil)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestGetClusterHealth(t *testing.T) {
	testCases := []struct {
		name     string
		response *corev2.HealthResponse
	}{
		{
			name:     "Healthy cluster",
			response: corev2.FixtureHealthResponse(true),
		},
		{
			name:     "Unhealthy cluster",
			response: corev2.FixtureHealthResponse(false),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewHealthController(store, nil, nil)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()

			// Mock store methods
			store.On("GetClusterHealth", ctx, nil, (*tls.Config)(nil)).Return(tc.response)

			// Exec GetClusterHealth
			response := actions.GetClusterHealth(ctx)

			// Assert
			assert.Equal(tc.response, response)
		})
	}
}
