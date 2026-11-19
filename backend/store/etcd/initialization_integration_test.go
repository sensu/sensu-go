//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keepAliveGoroutines() int {
	buf := make([]byte, 4<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), "concurrency.NewSession.func")
}

func TestInitializerLockFailureReleasesSession(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	s := NewStore(e.NewEmbeddedClient(), e.Name())

	holder, err := s.NewInitializer(context.Background())
	require.NoError(t, err)
	require.NoError(t, holder.Lock(context.Background()))
	defer func() { _ = holder.Close(context.Background()) }()

	before := keepAliveGoroutines()
	const n = 5
	for i := 0; i < n; i++ {
		waiter, err := s.NewInitializer(context.Background())
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		err = waiter.Lock(ctx)
		cancel()
		require.Error(t, err, "the lock is held, so Lock must fail")
	}

	var leaked int
	for i := 0; i < 100; i++ {
		leaked = keepAliveGoroutines() - before
		if leaked <= 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.LessOrEqual(t, leaked, 0, "session keepalive goroutines are still running after %d failed Lock calls", n)
}
