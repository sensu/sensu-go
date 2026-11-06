package agent

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openFDs(t *testing.T) int {
	dir := "/dev/fd"
	if runtime.GOOS == "linux" {
		dir = "/proc/self/fd"
	}
	d, err := os.Open(dir)
	require.NoError(t, err)
	defer d.Close()
	names, err := d.Readdirnames(-1)
	require.NoError(t, err)
	return len(names)
}

func TestExecuteHookAllowListDoesNotLeakFDs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no fd directory on windows")
	}
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	exe, err := filepath.Abs(os.Args[0])
	require.NoError(t, err)
	a.allowList = []allowList{{Exec: exe, Sha512: "not-the-real-sha"}}

	hookConfig := corev2.FixtureHookConfig("hook")
	hookConfig.Command = exe
	event := corev2.FixtureEvent("entity", "check")

	run := func() {
		hook := a.executeHook(context.Background(), hookConfig, event, nil)
		assert.NotNil(t, hook)
	}
	run() // warm up lazily opened fds
	before := openFDs(t)
	for i := 0; i < 50; i++ {
		run()
	}
	assert.LessOrEqual(t, openFDs(t), before+5, "executable opened for the allow-list check was not closed")
}
