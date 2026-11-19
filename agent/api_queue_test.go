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

func apiQueueInnerGoroutines() int {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), "(*Agent).handleAPIQueue.func1")
}

func TestHandleAPIQueueInnerGoroutineExitsOnCancel(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	// The agent is disconnected: sendq is full, so handleAPIQueue's consumer
	// blocks on its first message and the producer goroutine backs up behind it.
	for i := 0; i < cap(a.sendq); i++ {
		a.sendq <- &transport.Message{}
	}
	for i := 0; i < 4; i++ {
		_, err := a.apiQueue.Send(compressMessage([]byte("event")))
		require.NoError(t, err)
	}

	before := apiQueueInnerGoroutines()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { a.handleAPIQueue(ctx); close(done) }()
	require.Eventually(t, func() bool { return apiQueueInnerGoroutines()-before == 1 }, 2*time.Second, 10*time.Millisecond)

	// Give the producer time to fill the channel buffer and block handing off one more message.
	time.Sleep(300 * time.Millisecond)
	cancel()

	exited := func() bool { return apiQueueInnerGoroutines()-before <= 0 }
	ok := assertEventually(exited, 2*time.Second)

	// Unblock the consumer so the test leaves nothing behind.
	for len(a.sendq) > 0 {
		<-a.sendq
	}
	for i := 0; i < 20; i++ {
		select {
		case <-done:
			i = 20
		case <-time.After(50 * time.Millisecond):
			for len(a.sendq) > 0 {
				<-a.sendq
			}
		}
	}
	if !ok {
		t.Fatal("the API queue producer goroutine is still blocked handing a message to a consumer that has stopped")
	}
}

func assertEventually(cond func() bool, wait time.Duration) bool {
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}
func TestRunShutsDownWithQueuedAPIEventsWhileDisconnected(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.BackendURLs = []string{"ws://127.0.0.1:1"}
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	// The agent never connects, so nothing drains sendq: API events queued on
	// disk move into sendq, each carrying a callback that acks or nacks it.
	for i := 0; i < 40; i++ {
		_, err := a.apiQueue.Send(compressMessage([]byte("event")))
		require.NoError(t, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	require.Eventually(t, func() bool { return len(a.sendq) == cap(a.sendq) }, 5*time.Second, 10*time.Millisecond,
		"queued API events never filled sendq")
	cancel()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return: closing the API queue waits for events that were never acked or nacked")
	}
}

// A message taken by handleAPIQueue just before shutdown must not reach sendq
// after drainSendq has emptied it, or closing the queue blocks on that message.
func TestRunWaitsForAPIQueueLoopBeforeDrainingSendq(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.BackendURLs = []string{"ws://127.0.0.1:1"}
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	for i := 0; i < 40; i++ {
		_, err := a.apiQueue.Send(compressMessage([]byte("event")))
		require.NoError(t, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	require.Eventually(t, func() bool { return len(a.sendq) == cap(a.sendq) }, 5*time.Second, time.Millisecond)
	cancel()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return: a queued event reached sendq after it was drained")
	}
	// Nothing may be left in sendq once Run has returned.
	assert.Zero(t, len(a.sendq), "messages left in sendq after shutdown stay unacked")
}

// Run returns early when validation fails, before handleAPIQueue was started;
// waiting for it then would block forever.
func TestRunReturnsWhenValidationFailsBeforeAPIQueueStarts(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.BackendURLs = nil
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(context.Background()) }()
	select {
	case err := <-runDone:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Run blocked waiting for handleAPIQueue, which was never started")
	}
}
