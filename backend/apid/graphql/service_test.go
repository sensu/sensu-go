package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceSmokeTest(t *testing.T) {
	store := &mockstore.MockStore{}
	svc, err := NewService(ServiceConfig{Store: store, QueueGetter: queue.NewMemoryGetter()})
	require.NoError(t, err)
	assert.NotEmpty(t, svc)
}
