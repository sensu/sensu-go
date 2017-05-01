package transport

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testMessageType struct {
	Data string
}

func TestTransportSendReceive(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			done <- struct{}{}
		}()
		transport, err := Serve(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		msg, err := transport.Receive(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, "testMessageType", msg.Type)
		m := &testMessageType{"message"}
		assert.NoError(t, json.Unmarshal(msg.Payload, m))
		assert.Equal(t, testMessage.Data, m.Data)
	}))
	defer ts.Close()

	clientTransport, err := Connect(strings.Replace(ts.URL, "http", "ws", 1))
	assert.NoError(t, err)
	msgBytes, err := json.Marshal(testMessage)
	assert.NoError(t, err)
	err = clientTransport.Send(context.Background(), &Message{"testMessageType", msgBytes})
	assert.NoError(t, err)

	<-done
}

func TestClosedWebsocket(t *testing.T) {
	done := make(chan struct{}, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			done <- struct{}{}
		}()
		transport, err := Serve(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		transport.Close()
	}))
	defer ts.Close()

	clientTransport, err := Connect(strings.Replace(ts.URL, "http", "ws", 1))
	assert.NoError(t, err)
	<-done

	recvCtx, recvCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer recvCtxCancel()

	_, err = clientTransport.Receive(recvCtx)
	assert.Error(t, err)

	sendCtx, sendCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer sendCtxCancel()
	err = clientTransport.Send(sendCtx, &Message{"testMessageType", []byte{}})

	assert.Error(t, err)

	assert.Error(t, clientTransport.Error())
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
