package agent

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/transport"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apiQueueGoroutineName is how a stack dump names handleAPIQueue's producer
// goroutine. It depends on how the compiler numbers closures, so
// TestAPIQueueGoroutineMatcherSeesTheProducer checks it still matches.
const apiQueueGoroutineName = "(*Agent).handleAPIQueue.func1"

func apiQueueInnerGoroutines() int {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Count(string(buf), apiQueueGoroutineName)
}

// apiQueueProducerBlockedOnSend reports whether handleAPIQueue's producer
// goroutine is parked in its hand-off select, which only blocks once the
// channel buffer is full.
func apiQueueProducerBlockedOnSend() bool {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	for _, g := range strings.Split(string(buf), "\n\n") {
		if strings.Contains(g, apiQueueGoroutineName) && strings.Contains(g, "[select") {
			return true
		}
	}
	return false
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

	// Wait until the producer has filled the channel buffer and is blocked
	// handing off one more message, which is the state this test is about.
	require.Eventually(t, apiQueueProducerBlockedOnSend, 5*time.Second, 10*time.Millisecond)
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

// closeAfterFirstSend delegates to a real transport but closes it once the first
// message, the keepalive sendLoop sends before anything else, has been sent. That
// is the state the review described: the connection closes (shutdown, or
// receiveLoop hitting an error) while sendLoop is running and about to pick up a
// queued API event.
type closeAfterFirstSend struct {
	transport.Transport
	sent int32
}

func (c *closeAfterFirstSend) Send(m *transport.Message) error {
	err := c.Transport.Send(m)
	if atomic.AddInt32(&c.sent, 1) == 1 {
		_ = c.Transport.Close()
	}
	return err
}

// A queued API event taken by sendLoop after the transport closed must still be
// acked or nacked, or closing the API queue waits for it forever. This uses the
// real WebSocketTransport: a fake that ignores SendCallback would hide it.
func TestSendLoopNacksQueuedEventWhenTransportIsClosed(t *testing.T) {
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if conn, err := server.Serve(w, r); err == nil {
			for {
				if _, err := conn.Receive(); err != nil {
					return
				}
			}
		}
	}))
	defer ts.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http"), nil)
	require.NoError(t, err)
	conn := &closeAfterFirstSend{Transport: transport.NewTransport(ws)}

	a := newLoopTestAgent(t)
	_, err = a.apiQueue.Send(compressMessage([]byte("event")))
	require.NoError(t, err)
	queued, err := a.apiQueue.Receive(context.Background())
	require.NoError(t, err)
	a.sendq <- &transport.Message{
		Type:    transport.MessageTypeEvent,
		Payload: decompressMessage(queued.Body),
		SendCallback: func(err error) {
			if err != nil {
				_ = queued.Nack(true)
			} else {
				_ = queued.Ack()
			}
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sendDone := make(chan error, 1)
	go func() { sendDone <- a.sendLoop(ctx, cancel, conn) }()
	select {
	case err := <-sendDone:
		require.Error(t, err, "sendLoop must fail on the closed transport")
	case <-time.After(5 * time.Second):
		t.Fatal("sendLoop did not return")
	}

	closed := make(chan error, 1)
	go func() { closed <- a.apiQueue.Close() }()
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("closing the API queue blocked on an event that was taken from sendq and never acked or nacked")
	}
}

// Shutdown while the agent is connected: some queued events are sent, and the
// rest must be handed back so closing the queue does not block.
// Smoke test: shutting down while connected, with events flowing through
// sendLoop, returns promptly. The fixes themselves are guarded by
// TestSendCallbackRunsWhenTransportIsClosed (transport) and
// TestSendLoopNacksQueuedEventWhenTransportIsClosed; removing either makes
// those fail, whereas this test passes without them.
//
// Every queued event is an fsync'd write to the on-disk queue, which is slow on
// CI runners, so keep the count small: it only has to keep sendq busy.
func TestRunShutsDownWithQueuedAPIEventsWhileConnected(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()

	server := transport.NewServer()
	connected := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		if err != nil {
			return
		}
		select {
		case connected <- struct{}{}:
		default:
		}
		for {
			if _, err := conn.Receive(); err != nil {
				return
			}
		}
	}))
	defer ts.Close()

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	cfg.AgentManagedEntity = true
	cfg.BackendURLs = []string{"ws" + strings.TrimPrefix(ts.URL, "http")}
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- a.Run(ctx) }()
	select {
	case <-connected:
	case <-time.After(10 * time.Second):
		t.Fatal("agent never connected")
	}
	// Queue the events once the agent is connected. Doing it earlier makes
	// handleAPIQueue read the content type while connectWithBackoff is still
	// setting it, a pre-existing unsynchronised access that is not what this test
	// is about.
	require.Eventually(t, a.Connected, 10*time.Second, 5*time.Millisecond)
	for i := 0; i < 40; i++ {
		_, err := a.apiQueue.Send(compressMessage([]byte("event")))
		require.NoError(t, err)
	}
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-runDone:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return while connected with queued API events")
	}
}

// Shutdown hands every queued event back with context.Canceled. That is expected
// and must not log one error per event; a real send failure still does.
func TestQueuedEventCallbackLogLevel(t *testing.T) {
	hook := logrustest.NewGlobal()
	defer hook.Reset()
	old := logger.Logger.GetLevel()
	logger.Logger.SetLevel(logrus.DebugLevel)
	defer logger.Logger.SetLevel(old)

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	// Take one queued event through the real handleAPIQueue path and capture the
	// callback it builds.
	_, err = a.apiQueue.Send(compressMessage([]byte("event")))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go a.handleAPIQueue(ctx)
	var msg *transport.Message
	require.Eventually(t, func() bool {
		select {
		case msg = <-a.sendq:
			return true
		default:
			return false
		}
	}, 5*time.Second, 10*time.Millisecond)

	hook.Reset()
	msg.SendCallback(context.Canceled)
	for _, e := range hook.AllEntries() {
		assert.NotEqual(t, logrus.ErrorLevel, e.Level, "shutdown nack logged at error level: %q", e.Message)
	}
	assert.NotEmpty(t, hook.AllEntries(), "expected a debug line for the returned message")

	// A real failure is still an error, and the event still goes back on the queue.
	_, err = a.apiQueue.Send(compressMessage([]byte("event2")))
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		select {
		case msg = <-a.sendq:
			return true
		default:
			return false
		}
	}, 5*time.Second, 10*time.Millisecond)
	hook.Reset()
	msg.SendCallback(errors.New("write failed"))
	var sawError bool
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.ErrorLevel {
			sawError = true
		}
	}
	assert.True(t, sawError, "a real send failure must still be logged at error level")
}

// The goroutine matcher depends on how the compiler numbers closures. Prove that
// it still sees the producer goroutine, so a renamed closure cannot make the leak
// test pass without checking anything.
func TestAPIQueueGoroutineMatcherSeesTheProducer(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)

	before := apiQueueInnerGoroutines()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { a.handleAPIQueue(ctx); close(done) }()
	require.Eventually(t, func() bool { return apiQueueInnerGoroutines() > before }, 2*time.Second, 5*time.Millisecond,
		"apiQueueInnerGoroutines no longer matches handleAPIQueue's producer goroutine; update apiQueueGoroutineName")
	cancel()
	<-done
}
