package agent

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleTCPMessages(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
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

	payload := v1.CheckResult{
		Name:   "app_01",
		Output: "could not connect to something",
		Client: "proxEnt",
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
	assert.Equal("app_01", event.Check.Name)
	assert.Equal(int32(0), event.Check.Status)
	ta.Stop()
}

func TestHandleUDPMessages(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
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

	payload := v1.CheckResult{
		Name:   "app_01",
		Output: "could not connect to something",
		Client: "proxEnt",
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
	assert.Equal("app_01", event.Check.Name)
	assert.Equal(int32(0), event.Check.Status)
	ta.Stop()
}

func TestMultiWriteTimeoutTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
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

func TestReceiveMultiWriteTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
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

	payload := v1.CheckResult{
		Name:   "app_01",
		Output: "could not connect to something",
		Client: "proxEnt",
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
	assert.Equal("app_01", event.Check.Name)
	assert.Equal(int32(0), event.Check.Status)

	ta.Stop()
}

func TestReceivePingTCP(t *testing.T) {
	assert := assert.New(t)

	cfg := FixtureConfig()
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
