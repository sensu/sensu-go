package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewServiceSmokeTest(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On("NewQueue", mock.Anything, mock.Anything).Return(&mockqueue.MockQueue{})
	svc, err := NewService(ServiceConfig{Store: store})
	require.NoError(t, err)
	assert.NotEmpty(t, svc)
}
