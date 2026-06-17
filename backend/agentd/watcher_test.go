package agentd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
)

func TestGetEntityConfigWatcher(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	ch := GetEntityConfigWatcher(context.Background(), client)
	assert.NotNil(t, ch)
}

func TestGetUserConfigWatcher(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	ch := GetUserConfigWatcher(context.Background(), client)
	assert.NotNil(t, ch)
}
