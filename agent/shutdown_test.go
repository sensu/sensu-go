package agent

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With sendq full, handleAPIQueue's consumer is blocked in sendMessage holding
// an event it took from the on-disk queue. Closing the queue waits for that
// event to be acked or nacked, and sendMessage only gives it back once done is
// closed, so done must be closed before the queue is.
func TestRunReturnsWhenAnAPIEventIsBlockedInSendMessage(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.BackendURLs = []string{"ws://127.0.0.1:1"}
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}
	_, err = a.apiQueue.Send(compressMessage([]byte("event")))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	// Wait until handleAPIQueue's consumer has taken the event and is parked in
	// sendMessage: sendq is full, so nothing else can be holding it.
	require.Eventually(t, func() bool { return apiQueueConsumerBlockedInSendMessage() }, 10*time.Second, 5*time.Millisecond,
		"the API queue consumer never blocked in sendMessage")
	cancel()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return: the API queue is waiting for an event blocked in sendMessage, and done is only closed after it")
	}
}

// apiQueueConsumerBlockedInSendMessage reports whether a goroutine is parked in
// sendMessage on behalf of handleAPIQueue.
func apiQueueConsumerBlockedInSendMessage() bool {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	for _, g := range strings.Split(string(buf), "\n\n") {
		if strings.Contains(g, "(*Agent).sendMessage(") && strings.Contains(g, "(*Agent).handleAPIQueue(") && strings.Contains(g, "[select") {
			return true
		}
	}
	return false
}

type blockingPlugin struct {
	BasePlugin
	started chan struct{}
	release chan struct{}
}

func (p *blockingPlugin) BeforeAgentRun(ctx context.Context, a *Agent) error {
	close(p.started)
	return nil
}

func (p *blockingPlugin) Name() string  { return "blocking" }
func (p *blockingPlugin) Priority() int { return 0 }
func (p *blockingPlugin) IsFatal() bool { return false }
func (p *blockingPlugin) AfterAgentRun(ctx context.Context, a *Agent) error {
	<-p.release
	return nil
}

// Socket and statsd handlers wait in sendMessageWithin for up to 5s. They must
// be released as soon as shutdown begins, not only once Run has finished. A
// plugin holds Run in AfterAgentRun, so only the early release can free them.
func TestSendMessageWithinReturnsAsSoonAsShutdownBegins(t *testing.T) {
	// No mockTime here: the send timeout below must elapse on the real clock.
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.BackendURLs = []string{"ws://127.0.0.1:1"}
	a, err := NewAgent(cfg)
	require.NoError(t, err)
	plugin := &blockingPlugin{started: make(chan struct{}), release: make(chan struct{})}
	a.RegisterPlugin(plugin)

	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	select {
	case <-plugin.started:
	case <-time.After(10 * time.Second):
		t.Fatal("Run never started")
	}

	sent := make(chan bool, 1)
	go func() { sent <- a.sendMessageWithin(&transport.Message{}, 20*time.Second) }()
	// sendq is full and the agent is still running, so the call must still be
	// waiting. This only needs to hold for as long as it takes to check.
	select {
	case <-sent:
		t.Fatal("sendMessageWithin returned while the agent was still running and sendq was full")
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	cancelled := time.Now()
	select {
	case ok := <-sent:
		assert.False(t, ok)
		assert.Less(t, time.Since(cancelled), 3*time.Second)
	case <-time.After(5 * time.Second):
		t.Fatal("sendMessageWithin was not released when shutdown began; it waits for Run to finish")
	}
	close(plugin.release)
	require.NoError(t, <-runDone)
}
