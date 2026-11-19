package agent

import (
	"context"
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
	time.Sleep(500 * time.Millisecond) // let the consumer take the event and block in sendMessage
	cancel()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return: the API queue is waiting for an event blocked in sendMessage, and done is only closed after it")
	}
}

type blockingPlugin struct {
	BasePlugin
	release chan struct{}
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
	plugin := &blockingPlugin{release: make(chan struct{})}
	a.RegisterPlugin(plugin)

	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	time.Sleep(500 * time.Millisecond)

	sent := make(chan bool, 1)
	go func() { sent <- a.sendMessageWithin(&transport.Message{}, 20*time.Second) }()
	select {
	case <-sent:
		t.Fatal("sendMessageWithin returned while the agent was still running and sendq was full")
	case <-time.After(300 * time.Millisecond):
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
