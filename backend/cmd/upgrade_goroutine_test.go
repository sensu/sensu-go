package cmd

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// etcdClientGoroutines counts the gRPC client-side goroutines, which an etcd
// client that was never closed leaves running for each connection it made.
func etcdClientGoroutines() int {
	buf := make([]byte, 4<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), "grpc/internal/grpcsync.(*CallbackSerializer).run")
}

func TestUpgradeCommandClosesEtcdClient(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	// Other tests in this package leave a config file path in the environment
	// that no longer exists; UpgradeCommand would try to read it.
	t.Setenv("SENSU_BACKEND_CONFIG_FILE", "")
	viper.Reset()
	defer viper.Reset()
	var urls []string
	for _, addr := range e.GetClientURLs() {
		urls = append(urls, "http://"+addr)
	}
	viper.Set(flagEtcdClientURLs, urls)
	viper.Set(flagSkipConfirm, true)
	viper.Set(flagTimeout, 5)

	run := func() {
		cmd := UpgradeCommand()
		cmd.SetArgs(nil)
		require.NoError(t, cmd.RunE(cmd, nil))
	}
	run() // warm up lazily started goroutines
	time.Sleep(200 * time.Millisecond)
	before := etcdClientGoroutines()

	const n = 4
	for i := 0; i < n; i++ {
		run()
	}

	var leaked int
	for i := 0; i < 100; i++ {
		leaked = etcdClientGoroutines() - before
		if leaked <= 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.LessOrEqual(t, leaked, 0, "etcd client goroutines are still running after %d upgrade runs", n)
}
