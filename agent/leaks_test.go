package agent

import (
	"context"
	"net"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestSendMessageUnblocksOnShutdown(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}
	called := make(chan error, 1)
	returned := make(chan struct{})
	go func() {
		a.sendMessage(&transport.Message{SendCallback: func(err error) { called <- err }})
		close(returned)
	}()
	select {
	case <-returned:
		t.Fatal("sendMessage returned while the queue was full and the agent running")
	case <-time.After(100 * time.Millisecond):
	}
	a.doneOnce.Do(func() { close(a.done) })
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("sendMessage still blocked after the agent shut down")
	}
	assert.Error(t, <-called)
}

func TestEnforceMaxSessionLengthStopsWithContext(t *testing.T) {
	a := &Agent{maxSessionLength: time.Hour}
	ctx, cancel := context.WithCancel(context.Background())
	cancelled := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		a.enforceMaxSessionLength(ctx, func() { close(cancelled) })
		close(finished)
	}()
	cancel()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("enforceMaxSessionLength did not return when the connection context ended")
	}
	select {
	case <-cancelled:
		t.Fatal("session was cancelled by the timer instead of the context")
	default:
	}
}

func TestUDPListenerSurvivesInvalidDatagram(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, addr, err := a.createListenSockets(ctx)
	require.NoError(t, err)

	c, err := net.Dial("udp", addr)
	require.NoError(t, err)
	defer c.Close()
	_, err = c.Write([]byte("not json"))
	require.NoError(t, err)
	_, err = c.Write([]byte(`{"name":"app","output":"ok","source":"s","command":"c"}`))
	require.NoError(t, err)

	select {
	case msg := <-a.sendq:
		assert.Equal(t, transport.MessageTypeEvent, msg.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("UDP listener stopped after an invalid datagram")
	}
}

func TestPrometheusEndpointStopsWithAgent(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.PrometheusBinding = addr
	cfg.BackendURLs = []string{"ws://127.0.0.1:1"}
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()

	up := func() bool {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	}
	require.Eventually(t, up, 5*time.Second, 50*time.Millisecond, "prometheus endpoint never came up")
	cancel()
	require.NoError(t, <-runDone)
	assert.Eventually(t, func() bool { return !up() }, 3*time.Second, 50*time.Millisecond, "prometheus endpoint still listening after the agent stopped")
}

func TestSendMessageWithinGivesUpWhenQueueIsFull(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)
	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}
	start := time.Now()
	assert.False(t, a.sendMessageWithin(&transport.Message{}, 100*time.Millisecond))
	assert.Less(t, time.Since(start), 2*time.Second)

	<-a.sendq
	assert.True(t, a.sendMessageWithin(&transport.Message{}, 100*time.Millisecond))
}

func TestTCPSocketHandlerReleasesConnectionWhileDisconnected(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)
	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tcpAddr, _, err := a.createListenSockets(ctx)
	require.NoError(t, err)

	c, err := net.Dial("tcp", tcpAddr)
	require.NoError(t, err)
	defer c.Close()
	_, err = c.Write([]byte(`{"name":"app","output":"ok","source":"s","command":"c"}`))
	require.NoError(t, err)

	require.NoError(t, c.SetReadDeadline(time.Now().Add(socketSendTimeout+3*time.Second)))
	buf := make([]byte, 16)
	n, err := c.Read(buf)
	require.NoError(t, err, "socket handler never answered while the agent was disconnected")
	assert.Equal(t, "invalid", string(buf[:n]))
}

func TestCreateListenSocketsFailureLeavesNothingBehind(t *testing.T) {
	// Occupy a TCP port so that the TCP listen fails after the UDP one succeeded.
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer blocker.Close()
	port := blocker.Addr().(*net.TCPAddr).Port

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Host = "127.0.0.1"
	cfg.Socket.Port = port
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, err = a.createListenSockets(ctx)
	require.Error(t, err)

	waited := make(chan struct{})
	go func() { a.wg.Wait(); close(waited) }()
	select {
	case <-waited:
	case <-time.After(2 * time.Second):
		t.Fatal("waitgroup was incremented for listeners that never started, so shutdown would stall")
	}

	udp, err := net.ListenPacket("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	require.NoError(t, err, "the UDP listener was left open after the TCP listener failed")
	_ = udp.Close()
}

func TestNewAgentReleasesDatabasesWhenConstructionFails(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.AllowList = "allow-list.unsupported-extension"

	_, err := NewAgent(cfg)
	require.Error(t, err)

	for _, name := range []string{"assets.db", "queue.db"} {
		db, err := bolt.Open(filepath.Join(cfg.CacheDir, name), 0600, &bolt.Options{Timeout: 500 * time.Millisecond})
		require.NoError(t, err, "%s is still locked after NewAgent failed", name)
		_ = db.Close()
	}
}
