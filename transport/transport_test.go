package transport

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
		msgType, payload, err := transport.Receive(context.TODO())

		assert.NoError(t, err)
		assert.Equal(t, "testMessageType", msgType)
		m := &testMessageType{"message"}
		assert.NoError(t, json.Unmarshal(payload, m))
		assert.Equal(t, testMessage.Data, m.Data)
		done <- struct{}{}
	}))
	defer ts.Close()

	clientTransport, err := Connect(strings.Replace(ts.URL, "http", "ws", 1))
	assert.NoError(t, err)
	msgBytes, err := json.Marshal(testMessage)
	assert.NoError(t, err)
	err = clientTransport.Send(context.TODO(), "testMessageType", msgBytes)
	assert.NoError(t, err)

	<-done
}

func TestClosedWebsocket(t *testing.T) {
	done := make(chan struct{}, 1)

	server := NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transport, err := server.Serve(w, r)
		assert.NoError(t, err)
		transport.Connection.Close()
		done <- struct{}{}
	}))
	defer ts.Close()

	clientTransport, err := Connect(strings.Replace(ts.URL, "http", "ws", 1))
	assert.NoError(t, err)
	<-done
	_, _, err = clientTransport.Receive(context.TODO())
	assert.IsType(t, ConnectionError{}, err)

	// This test will fail until https://github.com/gorilla/websocket/issues/226
	// is fixed. The first call to Send() will fail silently because of a bug
	// in the websocket library. Nothing we can do to prevent lost messages
	// right now.
	// err = clientTransport.Send(context.TODO(), "type", []byte("message"))
	// assert.IsType(t, ConnectionError{}, err)
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
