package agent

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		msg, err := conn.Receive()
		assert.NoError(t, err)
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
	cfg.Socket.Port = 0
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
	cfg.Socket.Port = 0
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

func TestHandleTCPMessages(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfig()
	// Assign a random port to the socket to avoid overlaps
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)

	addr, _, err := ta.createListenSockets()
	assert.NoError(err)
	if err != nil {
		assert.FailNow("createListenSockets() failed to run")
	}

	tcpClient, err := net.Dial("tcp", addr)
	if err != nil {
		assert.FailNow("failed to create TCP connection")
	}

	payload := map[string]interface{}{
		"timestamp": int64(1513028652),
		"name":      "app_01",
		"output":    "could not connect to something",
		"status":    1,
		"custom":    "attribute",
		"source":    "proxEnt",
	}
	bytes, _ := json.Marshal(payload)

	_, err = tcpClient.Write(bytes)
	require.NoError(t, err)
	require.NoError(t, tcpClient.Close())

	msg := <-ta.sendq
	assert.NotEmpty(msg)
	assert.Equal("event", msg.Type)

	var event types.Event
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		assert.FailNow("failed to unmarshal event json")
	}

	assert.NotNil(event.Entity)
	assert.Equal(int64(1513028652), event.Timestamp)
	assert.Equal("app_01", event.Check.Config.Name)
	ta.Stop()
}

func TestHandleUDPMessages(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfig()
	// Assign a random port to the socket to avoid overlaps
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)

	_, addr, err := ta.createListenSockets()
	assert.NoError(err)
	if err != nil {
		assert.FailNow("createListenSockets() failed to run")
	}

	udpClient, err := net.Dial("udp", addr)
	if err != nil {
		assert.FailNow("failed to create UDP connection")
	}

	payload := map[string]interface{}{
		"timestamp": int64(1513028652),
		"name":      "app_01",
		"output":    "could not connect to something",
		"status":    1,
		"custom":    "attribute",
		"source":    "proxEnt",
	}
	bytes, _ := json.Marshal(payload)

	_, err = udpClient.Write(bytes)
	require.NoError(t, err)
	require.NoError(t, udpClient.Close())

	msg := <-ta.sendq
	assert.NotEmpty(msg)
	assert.Equal("event", msg.Type)

	var event types.Event
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		assert.FailNow("Failed to unmarshal event json")
	}

	assert.NotNil(event.Entity)
	assert.Equal(int64(1513028652), event.Timestamp)
	assert.Equal("app_01", event.Check.Config.Name)
	ta.Stop()
}

func TestReceivePingTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfig()
	// Assign a random port to the socket to avoid overlaps
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)

	addr, _, err := ta.createListenSockets()
	assert.NoError(err)
	if err != nil {
		assert.FailNow("createListenSockets() failed to run")
	}

	tcpClient, err := net.Dial("tcp", addr)
	if err != nil {
		assert.FailNow("failed to create TCP connection")
	}
	defer func() {
		require.NoError(t, tcpClient.Close())
	}()

	bytesWritten, err := tcpClient.Write([]byte(" ping "))
	if err != nil {
		assert.FailNow("Failed to write to tcp server %s", err)
	}
	assert.Equal(6, bytesWritten)

	readData := make([]byte, 4)
	numBytes, err := tcpClient.Read(readData)
	require.NoError(t, err)
	assert.Equal("pong", string(readData[:numBytes]))
	ta.Stop()
}

func TestReceiveMultiWriteTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfig()
	// Assign a random port to the socket to avoid overlaps
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)

	addr, _, err := ta.createListenSockets()
	assert.NoError(err)
	if err != nil {
		assert.FailNow("createListenSockets() failed to run")
	}

	tcpClient, err := net.Dial("tcp", addr)
	if err != nil {
		assert.FailNow("failed to create TCP connection")
	}

	payload := map[string]interface{}{
		"timestamp": int64(1513028652),
		"name":      "app_01",
		"output":    "could not connect to something",
		"status":    1,
		"custom":    "attribute",
		"source":    "proxEnt",
	}
	bytes, _ := json.Marshal(payload)

	_, err = tcpClient.Write(bytes[:5])
	require.NoError(t, err)
	_, err = tcpClient.Write(bytes[5:])
	require.NoError(t, err)
	require.NoError(t, tcpClient.Close())

	msg := <-ta.sendq
	assert.Equal("event", msg.Type)

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotNil(event.Entity)
	assert.Equal(int64(1513028652), event.Timestamp)
	assert.Equal("app_01", event.Check.Config.Name)

	ta.Stop()
}

func TestMultiWriteTimeoutTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := NewConfig()
	// Assign a random port to the socket to avoid overlaps
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)

	addr, _, err := ta.createListenSockets()
	assert.NoError(err)
	if err != nil {
		assert.FailNow("createListenSockets() failed to run")
	}

	var checkString string
	for i := 0; i < 1500; i++ {
		checkString += "a"
	}

	chunkData := []byte(`{"timestamp":123, "check":{"output": "` + checkString + `"}}`)
	tcpClient, err := net.Dial("tcp", addr)
	if err != nil {
		assert.FailNow("failed to create TCP connection")
	}

	_, err = tcpClient.Write(chunkData[:5])
	if err != nil {
		assert.FailNow("failed to write data to tcp socket")
	}
	readData := make([]byte, 7)

	numBytes, err := bufio.NewReader(tcpClient).Read(readData)
	if err != nil {
		assert.FailNow("Failed to read data from tcp socket")
	}
	assert.Equal("invalid", string(readData[:numBytes]))
	require.NoError(t, tcpClient.Close())
	ta.Stop()
}
