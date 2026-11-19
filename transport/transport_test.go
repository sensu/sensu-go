package transport

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMessageType struct {
	Data string
}

func TestTransportSendReceive(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transport, err := server.Serve(w, r)
		assert.NoError(t, err)
		msg, err := transport.Receive()

		assert.NoError(t, err)
		assert.Equal(t, "testMessageType", msg.Type)
		m := &testMessageType{"message"}
		assert.NoError(t, json.Unmarshal(msg.Payload, m))
		assert.Equal(t, testMessage.Data, m.Data)
		done <- struct{}{}
	}))
	defer ts.Close()

	clientTransport, _, err := Connect(strings.Replace(ts.URL, "http", "ws", 1), nil, nil, 5)
	assert.NoError(t, err)
	msgBytes, err := json.Marshal(testMessage)
	assert.NoError(t, err)
	err = clientTransport.Send(&Message{
		Type:    "testMessageType",
		Payload: msgBytes,
	})
	assert.NoError(t, err)

	<-done
}

func TestClosedWebsocket(t *testing.T) {
	done := make(chan struct{}, 1)

	server := NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transport, err := server.Serve(w, r)
		assert.NoError(t, err)
		require.NoError(t, transport.Close())
		done <- struct{}{}
	}))
	defer ts.Close()

	clientTransport, _, err := Connect(strings.Replace(ts.URL, "http", "ws", 1), nil, nil, 5)
	assert.NoError(t, err)
	<-done
	// At this point we should receive a connection closed message.
	_, err = clientTransport.Receive()
	assert.IsType(t, ClosedError{}, err)

	err = clientTransport.Send(&Message{
		Type:    "testMessageType",
		Payload: []byte{}},
	)
	assert.IsType(t, ClosedError{}, err)

	_, err = clientTransport.Receive()
	assert.IsType(t, ClosedError{}, err)
}

// This was all mostly to prove that performance of encoding/decoding was
// not super-linear.

var (
	encodingTestMessages = map[int][]byte{}
)

func init() {
	sizes := []int{32, 64, 128, 1024, 32 * 1024, 128 * 1024}
	for _, sz := range sizes {
		encodingTestMessages[sz] = makeMessage(sz)
	}
}

func makeMessage(i int) []byte {
	msg := make([]byte, i)
	_, err := rand.Read(msg)
	if err != nil {
		panic(err)
	}
	return msg
}

func benchmarkEncode(i int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		Encode("type", encodingTestMessages[i])
	}
}

func BenchmarkEncode32(b *testing.B) {
	benchmarkEncode(32, b)
}

func BenchmarkEncode64(b *testing.B) {
	benchmarkEncode(64, b)
}

func BenchmarkEncode128(b *testing.B) {
	benchmarkEncode(128, b)
}

func BenchmarkEncode1k(b *testing.B) {
	benchmarkEncode(1024, b)
}

func BenchmarkEncode32k(b *testing.B) {
	benchmarkEncode(32*1024, b)
}

func BenchmarkEncode128k(b *testing.B) {
	benchmarkEncode(128*1024, b)
}

type closeTrackingConn struct {
	net.Conn
	closed int32
}

func (c *closeTrackingConn) Close() error {
	atomic.StoreInt32(&c.closed, 1)
	return c.Conn.Close()
}

func TestCloseAfterSendError(t *testing.T) {
	server := NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := server.Serve(w, r); err != nil {
			return
		}
	}))
	defer ts.Close()

	var tracked *closeTrackingConn
	dialer := websocket.Dialer{
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := (&net.Dialer{}).DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			tracked = &closeTrackingConn{Conn: c}
			return tracked, nil
		},
	}
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	tr := NewTransport(conn)

	// Kill the network under the websocket, bypassing the tracking wrapper.
	require.NoError(t, tracked.Conn.Close())

	require.Error(t, tr.Send(NewMessage("event", []byte("x"))))
	_ = tr.Close()

	assert.EqualValues(t, 1, atomic.LoadInt32(&tracked.closed), "underlying connection was not closed after a Send error")
}
