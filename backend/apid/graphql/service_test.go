package graphql

import (
	"testing"

	client "github.com/sensu/sensu-go/backend/apid/graphql/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceSmokeTest(t *testing.T) {
	_, factory := client.NewClientFactory()
	svc, err := NewService(ServiceConfig{ClientFactory: factory})
	require.NoError(t, err)
	assert.NotEmpty(t, svc)
}
