package agentd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/stretchr/testify/assert"
)

func TestGetEntityConfigWatcher(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()
	store := etcdstore.NewStore(client)

	ch := GetEntityConfigWatcher(context.Background(), store)
	assert.NotNil(t, ch)
}
