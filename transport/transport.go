package transport

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	sep = []byte("\n")
)

// A ClosedError is returned when Receive or Send is called on a closed
// Transport.
type ClosedError struct {
	Message string
}

func (e ClosedError) Error() string {
	return fmt.Sprintf("Connection closed: %s", e.Message)
}

// A ConnectionError is returned when a Transport receives any unexpected error
// connecting to, sending to, or receiving from a backend.
type ConnectionError struct {
	Message string
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("Connection error: %s", e.Message)
}

// Encode a message to be sent over a websocket channel
func Encode(msgType string, payload []byte) []byte {
	buf := []byte(msgType + "\n")
	buf = append(buf, payload...)
	return buf
}

// Decode a message received from a websocket channel.
func Decode(payload []byte) (string, []byte, error) {
	nl := bytes.Index(payload, sep)
	if nl < 0 {
		return "", nil, errors.New("invalid message")
	}

	msgType := payload[0:nl]
	msg := payload[nl+1:]
	return string(msgType), msg, nil
}

// A Message is a tuple of a message type (i.e. channel) and a byte-array
// payload to be sent across the transport.
type Message struct {
	Type    string
	Payload []byte
}

// A Transport is a connection between sensu Agents and Backends.
type Transport struct {
	Connection *websocket.Conn
	closed     bool
	mutex      *sync.RWMutex
}

// NewTransport creates an initialized Transport and return its pointer.
func NewTransport(conn *websocket.Conn) *Transport {
	return &Transport{
		Connection: conn,
		closed:     false,
		mutex:      &sync.RWMutex{},
	}
}

// Closed returns true if the underlying connection is closed.
func (t *Transport) Closed() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.closed
}

// Send is used to send a message over the transport. It takes a message type
// hint and a serialized payload. Send will block until the message has been
// sent. Send is synchronous, returning nil if the write to the underlying
// socket was successful and an error otherwise.
func (t *Transport) Send(m *Message) error {
	t.mutex.RLock()
	if t.closed {
		t.mutex.RUnlock()
		return ClosedError{"connection closed"}
	}
	t.mutex.RUnlock()

	msg := Encode(m.Type, m.Payload)
	err := t.Connection.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		// If we get _any_ error, let's just considered the connection closed,
		// because it's _really_ hard to figure out what errors from the
		// websocket library are terminal and which aren't. So, abandon all
		// hope, and reconnect if we get an error from the websocket lib.
		t.mutex.Lock()
		t.closed = true
		t.mutex.Unlock()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return ClosedError{err.Error()}
		}
		return ConnectionError{err.Error()}
	}

	return nil
}

// Receive is used to receive a message from the transport. It takes a context
// and blocks until the next message is received from the transport.
func (t *Transport) Receive() (*Message, error) {
	t.mutex.RLock()
	if t.closed {
		t.mutex.RUnlock()
		return nil, ClosedError{"connection closed"}
	}
	t.mutex.RUnlock()

	_, p, err := t.Connection.ReadMessage()
	if err != nil {
		t.mutex.Lock()
		t.closed = true
		t.mutex.Unlock()

		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return nil, ClosedError{err.Error()}
		}
		return nil, ConnectionError{err.Error()}
	}

	msgType, payload, err := Decode(p)
	if err != nil {
		return nil, err
	}

	return &Message{msgType, payload}, nil
}

// Close will cleanly shutdown a websocket connection.
func (t *Transport) Close() error {
	return t.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "bye"))
}
