package sender

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ash2k/stager/wait"
	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	t.Parallel()
	dc := dummyConn{}
	sender := Sender{
		ConnFactory: func() (net.Conn, error) {
			return &dc, nil
		},
		Sink: make(chan Stream),
		BufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	var wg wait.Group
	defer wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wg.StartWithContext(ctx, sender.Run)
	var wgTest sync.WaitGroup
	for i := 0; i <= 4; i++ {
		wgTest.Add(1)
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			sink := make(chan *bytes.Buffer, i)
			sender.Sink <- Stream{
				Cb: func(errs []error) {
					defer wgTest.Done()
					for _, e := range errs {
						assert.NoError(t, e)
					}
				},
				Buf: sink,
			}
			for x := 0; x < i; x++ {
				sink <- bytes.NewBuffer([]byte{byte(x)})
			}
			close(sink)
		})
	}
	wgTest.Wait()
	assert.Equal(t, []byte{0x0, 0x0, 0x1, 0x0, 0x1, 0x2, 0x0, 0x1, 0x2, 0x3}, dc.buf.Bytes())
}

func TestSendCallsCallbacksOnMainCtxDone(t *testing.T) {
	t.Parallel()
	sender := Sender{
		ConnFactory: func() (net.Conn, error) {
			return nil, errors.New("(donotwant)")
		},
		Sink: make(chan Stream, 1),
		BufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	var wg wait.Group
	defer wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wg.StartWithContext(ctx, sender.Run)
	var cbWg sync.WaitGroup
	cbWg.Add(1)
	sender.Sink <- Stream{
		Ctx: context.Background(),
		Cb: func(errs []error) {
			defer cbWg.Done()
			if assert.Len(t, errs, 1) {
				assert.Equal(t, context.Canceled, errs[0])
			}
		},
		Buf: make(<-chan *bytes.Buffer),
	}
	cancel()
	cbWg.Wait()
}

func TestSendCallsCallbackOnCtxDone1(t *testing.T) {
	t.Parallel()
	sender := Sender{
		ConnFactory: func() (net.Conn, error) {
			return nil, errors.New("(donotwant)")
		},
		Sink: make(chan Stream, 2),
		BufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	var wg wait.Group
	defer wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wg.StartWithContext(ctx, sender.Run)
	var cbWg sync.WaitGroup
	cbWg.Add(2)
	ctx1, cancel1 := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel1()
	sender.Sink <- Stream{
		Ctx: ctx1,
		Cb: func(errs []error) {
			defer cbWg.Done()
			if assert.Len(t, errs, 1) {
				assert.Equal(t, context.DeadlineExceeded, errs[0])
			}
		},
		Buf: make(<-chan *bytes.Buffer),
	}
	ctx2, cancel2 := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel2()
	sender.Sink <- Stream{
		Ctx: ctx2,
		Cb: func(errs []error) {
			defer cbWg.Done()
			if assert.Len(t, errs, 1) {
				assert.Equal(t, context.DeadlineExceeded, errs[0])
			}
		},
		Buf: make(<-chan *bytes.Buffer),
	}
	cbWg.Wait()
}

func TestSendCallsCallbackOnCtxDone2(t *testing.T) {
	t.Parallel()
	getFail := false
	sender := Sender{
		ConnFactory: func() (net.Conn, error) {
			if getFail {
				return nil, errors.New("(donotwant)")
			}
			getFail = true
			return &dummyConn{
				writeErr: errors.New("write fail"),
			}, nil
		},
		Sink: make(chan Stream, 2),
		BufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	var wg wait.Group
	defer wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wg.StartWithContext(ctx, sender.Run)
	var cbWg sync.WaitGroup
	cbWg.Add(2)
	ctx1, cancel1 := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel1()
	buf := make(chan *bytes.Buffer, 1)
	buf <- &bytes.Buffer{}
	close(buf)
	sender.Sink <- Stream{
		Ctx: ctx1,
		Cb: func(errs []error) {
			defer cbWg.Done()
			if assert.Len(t, errs, 2) {
				assert.EqualError(t, errs[0], "write fail")
				assert.Equal(t, context.DeadlineExceeded, errs[1])
			}
		},
		Buf: buf,
	}
	ctx2, cancel2 := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel2()
	sender.Sink <- Stream{
		Ctx: ctx2,
		Cb: func(errs []error) {
			defer cbWg.Done()
			if assert.Len(t, errs, 1) {
				assert.Equal(t, context.DeadlineExceeded, errs[0])
			}
		},
		Buf: make(chan *bytes.Buffer),
	}
	cbWg.Wait()
}

type dummyConn struct {
	buf      bytes.Buffer
	writeErr error
	isClosed bool
}

func (c *dummyConn) Read(b []byte) (int, error) {
	return 0, errors.New("asdasd")
}

func (c *dummyConn) Write(b []byte) (int, error) {
	if c.isClosed {
		panic("closed")
	}
	if c.writeErr != nil {
		return 0, c.writeErr
	}
	return c.buf.Write(b)
}

func (c *dummyConn) Close() error {
	c.isClosed = true
	return nil
}

func (c *dummyConn) LocalAddr() net.Addr {
	return nil
}

func (c *dummyConn) RemoteAddr() net.Addr {
	return nil
}

func (c *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}
