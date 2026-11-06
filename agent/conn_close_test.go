package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTransport struct {
	sendErr    error
	receiveErr error
	closeCalls int32
	closeCh    chan struct{}
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{closeCh: make(chan struct{}, 8)}
}

func (f *fakeTransport) Close() error {
	atomic.AddInt32(&f.closeCalls, 1)
	f.closeCh <- struct{}{}
	return nil
}
func (f *fakeTransport) Closed() bool                         { return atomic.LoadInt32(&f.closeCalls) > 0 }
func (f *fakeTransport) Heartbeat(context.Context, int, int)  {}
func (f *fakeTransport) Send(*transport.Message) error        { return f.sendErr }
func (f *fakeTransport) SendCloseMessage() error              { return nil }
func (f *fakeTransport) Receive() (*transport.Message, error) { return nil, f.receiveErr }

func newLoopTestAgent(t *testing.T) *Agent {
	cfg, cleanup := FixtureConfig()
	t.Cleanup(cleanup)
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	a, err := NewAgent(cfg)
	require.NoError(t, err)
	return a
}

func TestSendLoopClosesConnOnSendError(t *testing.T) {
	a := newLoopTestAgent(t)
	conn := newFakeTransport()
	conn.sendErr = errors.New("write failed")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := a.sendLoop(ctx, cancel, conn)
	require.Error(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&conn.closeCalls))
}

func TestSendLoopClosesConnOnContextDone(t *testing.T) {
	a := newLoopTestAgent(t)
	conn := newFakeTransport()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, a.sendLoop(ctx, cancel, conn))
	assert.EqualValues(t, 1, atomic.LoadInt32(&conn.closeCalls))
}

func TestReceiveLoopClosesConnOnReceiveError(t *testing.T) {
	a := newLoopTestAgent(t)
	conn := newFakeTransport()
	conn.receiveErr = errors.New("read failed")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.receiveLoop(ctx, cancel, conn)
	assert.EqualValues(t, 1, atomic.LoadInt32(&conn.closeCalls))
}

func TestReceiveLoopClosesConnOnContextDone(t *testing.T) {
	a := newLoopTestAgent(t)
	conn := newFakeTransport()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a.receiveLoop(ctx, cancel, conn)
	assert.EqualValues(t, 1, atomic.LoadInt32(&conn.closeCalls))
}
