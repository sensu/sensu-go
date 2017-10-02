package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testMessageType struct {
	Data string
}

func TestSendLoop(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()
		msg, err := conn.Receive()

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "keepalive", msg.Type)
		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.NotNil(t, event.Entity)
		assert.Equal(t, "agent", event.Entity.Class)
		assert.NotEmpty(t, event.Entity.System.Hostname)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	<-done
	ta.Stop()
}

func TestReceiveLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()

		msgBytes, err := json.Marshal(testMessage)
		assert.NoError(t, err)
		tm := &transport.Message{
			Type:    "testMessageType",
			Payload: msgBytes,
		}
		err = conn.Send(tm)
		assert.NoError(t, err)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	ta := NewAgent(cfg)
	ta.addHandler("testMessageType", func(payload []byte) error {
		msg := &testMessageType{}
		err := json.Unmarshal(payload, msg)
		assert.NoError(t, err)
		assert.Equal(t, testMessage.Data, msg.Data)
		done <- struct{}{}
		return nil
	})
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	ta.sendMessage("testMessageType", msgBytes)
	<-done
	<-done
	ta.Stop()
}

func TestReconnect(t *testing.T) {
	control := make(chan struct{})
	connectionCount := 0
	server := transport.NewServer()
	mutex := &sync.Mutex{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()
		mutex.Lock()
		connectionCount++
		mutex.Unlock()
		<-control
		conn.Close()
	}))
	defer ts.Close()

	// connect with an agent
	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	control <- struct{}{}
	mutex.Lock()
	assert.Equal(t, 1, connectionCount)
	mutex.Unlock()

	control <- struct{}{}
	mutex.Lock()
	assert.Condition(t, func() bool { return connectionCount > 1 })
	mutex.Unlock()
	ta.Stop()
}

func TestReceiveLoopTCP(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive

		msg, err := conn.Receive() // our message

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "event", msg.Type)
		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.Equal(t, int64(123), event.Timestamp)
		assert.NotNil(t, event.Entity)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	tcpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create TCP connection")
	}

	defer tcpClient.Close()

	tcpClient.Write([]byte(`{"timestamp":123}`))
	tcpClient.Close()
	<-done
	ta.Stop()
}

func TestReceiveLoopCheckTCP(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive

		msg, err := conn.Receive() // our message

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "event", msg.Type)
		event := &types.Event{}
		check := &types.Check{Status: 1}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.Equal(t, int64(123), event.Timestamp)
		assert.Equal(t, check, event.Check)
		assert.NotNil(t, event.Entity)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	tcpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create TCP connection")
	}

	defer tcpClient.Close()

	tcpClient.Write([]byte(`{"timestamp":123, "check":{"status":1}}`))
	tcpClient.Close()
	<-done
	ta.Stop()
}

func TestReceiveLoopUDP(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive

		msg, err := conn.Receive() // our message

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "event", msg.Type)

		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.Equal(t, int64(123), event.Timestamp)
		assert.NotNil(t, event.Entity)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	udpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create UDP connection")
	}
	defer udpClient.Close()

	udpClient.Write([]byte(`{"timestamp":123}`))
	udpClient.Close()
	<-done
	ta.Stop()
}

func TestReceiveLoopPing(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive

		close(done)
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	readData := make([]byte, 4)
	tcpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create TCP connection")
	}
	defer tcpClient.Close()
	bytesWritten, err := tcpClient.Write([]byte(" ping "))
	if err != nil {
		assert.FailNow(t, "Failed to write to tcp server %s", err)
	}
	assert.Equal(t, 6, bytesWritten)
	numBytes, err := tcpClient.Read(readData)
	if err != nil {
		fmt.Println(err)
		assert.FailNow(t, "failed to read tcpClient")
	}
	assert.Equal(t, "pong", string(readData[:numBytes]))
	tcpClient.Close()
	<-done
	ta.Stop()
}

func TestReceiveLoopMultiWriteTCP(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive

		msg, err := conn.Receive() // our message

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "event", msg.Type)
		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.Equal(t, int64(123), event.Timestamp)
		assert.NotNil(t, event.Entity)
		close(done)
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	var checkString string
	for i := 0; i < 1500; i++ {
		checkString += "a"
	}

	chunkData := []byte(`{"timestamp":123, "check":{"output": "` + checkString + `"}}`)
	tcpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create TCP connection")
	}
	tcpClient.Write(chunkData[:5])
	tcpClient.Write(chunkData[5:])
	tcpClient.Close()
	<-done
	ta.Stop()
}

func TestReceiveLoopMultiWriteTimeoutTCP(t *testing.T) {

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive() // agent handshake
		conn.Receive() // agent keepalive
		close(done)
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURLs = []string{wsURL}
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}

	var checkString string
	for i := 0; i < 1500; i++ {
		checkString += "a"
	}

	chunkData := []byte(`{"timestamp":123, "check":{"output": "` + checkString + `"}}`)
	tcpClient, err := net.Dial("tcp", ":3030")
	if err != nil {
		assert.FailNow(t, "failed to create TCP connection")
	}

	_, err = tcpClient.Write(chunkData[:5])
	if err != nil {
		assert.FailNow(t, "failed to write data to tcp socket")
	}
	readData := make([]byte, 7)

	numBytes, err := bufio.NewReader(tcpClient).Read(readData)
	if err != nil {
		assert.FailNow(t, "Failed to read data from tcp socket")
	}
	assert.Equal(t, "invalid", string(readData[:numBytes]))
	tcpClient.Close()
	<-done
	ta.Stop()
}
