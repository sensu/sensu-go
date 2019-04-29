// +build integration,!race

package limitd

import (
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/limiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLimitdTest(t *testing.T) *Limitd {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	limitd, err := New(Config{
		Client:        client,
		EntityLimiter: limiter.NewEntityLimiter(),
	})
	require.NoError(t, err)
	return limitd
}

func TestLimitd(t *testing.T) {
	limitd := newLimitdTest(t)
	require.NoError(t, limitd.Start())
	assert.NoError(t, limitd.Stop())
	assert.Equal(t, limitd.Name(), componentName)
}
