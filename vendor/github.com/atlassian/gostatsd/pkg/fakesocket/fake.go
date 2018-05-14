package fakesocket

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

// FakeMetric is a fake metric.
var FakeMetric = []byte("foo.bar.baz:2|c")

// FakeAddr is a fake net.Addr
var FakeAddr = &net.UDPAddr{
	IP:   net.IPv4(127, 0, 0, 1),
	Port: 8181,
}

var ErrClosedConnection = errors.New("Connection is closed")
var ErrAlreadyClosedConnection = errors.New("Connection is already closed")

// FakePacketConn is a fake net.PacketConn (and net.Conn) providing FakeMetric when read from.
type FakePacketConn struct {
	closedFlag uint64
	closedChan chan struct{}
	count      uint64
	limit      uint64
}

func (fpc *FakePacketConn) isClosed() bool {
	return atomic.LoadUint64(&fpc.closedFlag) != 0
}

// ReadFrom copies FakeMetric into b.
func (fpc *FakePacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if fpc.isClosed() {
		return 0, nil, ErrClosedConnection
	}

	if atomic.AddUint64(&fpc.count, 1) >= fpc.limit {
		_ = fpc.Close() // nolint:gas
		return 0, nil, ErrClosedConnection
	}

	n := copy(b, FakeMetric)
	return n, FakeAddr, nil
}

// WriteTo dummy impl.
func (fpc *FakePacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if fpc.isClosed() {
		return 0, ErrClosedConnection
	}
	return 0, nil
}

// Close dummy impl.
func (fpc *FakePacketConn) Close() error {
	if atomic.SwapUint64(&fpc.closedFlag, 1) == 0 {
		// Was open, now closed
		close(fpc.closedChan)
		return nil
	} else {
		// Already closed
		return ErrAlreadyClosedConnection
	}
}

// LocalAddr dummy impl.
func (fpc *FakePacketConn) LocalAddr() net.Addr { return FakeAddr }

// SetDeadline dummy impl.
func (fpc *FakePacketConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline dummy impl.
func (fpc *FakePacketConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline dummy impl.
func (fpc *FakePacketConn) SetWriteDeadline(t time.Time) error { return nil }

// Read copies FakeMetric into b
// To satisfy net.Conn
func (fpc *FakePacketConn) Read(b []byte) (int, error) {
	n, _, err := fpc.ReadFrom(b)
	return n, err
}

// WriteTo dummy impl.
// To satisfy net.Conn
func (fpc *FakePacketConn) Write(b []byte) (int, error) {
	return fpc.WriteTo(b, FakeAddr)
}

// RemoteAddr dummy impl.
// To satisfy net.Conn
func (fpc *FakePacketConn) RemoteAddr() net.Addr { return FakeAddr }

// FakeRandomPacketConn is a fake net.PacketConn providing random fake metrics.
type FakeRandomPacketConn struct {
	FakePacketConn
}

// ReadFrom generates random metric and writes in into b.
func (frpc *FakeRandomPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if frpc.isClosed() {
		return 0, nil, ErrClosedConnection
	}

	num := rand.Int31n(10000) // Randomize metric name
	buf := new(bytes.Buffer)
	switch rand.Int31n(4) {
	case 0: // Counter
		fmt.Fprintf(buf, "statsd.tester.counter_%d:%f|c\n", num, rand.Float64()*100) // #nosec
	case 1: // Gauge
		fmt.Fprintf(buf, "statsd.tester.gauge_%d:%f|g\n", num, rand.Float64()*100) // #nosec
	case 2: // Timer
		n := 10
		for i := 0; i < n; i++ {
			fmt.Fprintf(buf, "statsd.tester.timer_%d:%f|ms\n", num, rand.Float64()*100) // #nosec
		}
	case 3: // Set
		for i := 0; i < 10; i++ {
			fmt.Fprintf(buf, "statsd.tester.set_%d:%d|s\n", num, rand.Int31n(9)+1) // #nosec
		}
	default:
		panic(errors.New("unreachable"))
	}
	n := copy(b, buf.Bytes())
	return n, FakeAddr, nil
}

// Factory is a replacement for net.ListenPacket() that produces instances of FakeRandomPacketConn.
func Factory() (net.PacketConn, error) {
	frpc := &FakeRandomPacketConn{
		FakePacketConn: FakePacketConn{
			closedChan: make(chan struct{}),
		},
	}
	return frpc, nil
}

func NewFakePacketConn() net.PacketConn {
	return &FakePacketConn{
		limit:      math.MaxInt64, // not Uint64, it needs room to overflow without wrapping
		closedChan: make(chan struct{}),
	}
}

func NewCountedFakePacketConn(limit uint64) (net.PacketConn, chan struct{}) {
	ch := make(chan struct{})
	return &FakePacketConn{
		limit:      limit,
		closedChan: ch,
	}, ch
}
