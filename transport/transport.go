package transport

import (
	"bytes"
	"errors"
	"fmt"

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

// A Transport is a connection between sensu Agents and Backends.
type Transport struct {
	Connection *websocket.Conn
}

// NewTransport creates an initialized Transport and return its pointer.
func NewTransport(conn *websocket.Conn) *Transport {
	return &Transport{
		Connection: conn,
	}
}

// Send is used to send a message over the transport. It takes a message type
// hint and a serialized payload. Send will block until the message has been
// sent. Send is synchronous, returning nil if the write to the underlying
// socket was successful and an error otherwise.
func (t *Transport) Send(msgType string, payload []byte) error {
	msg := Encode(msgType, payload)
	err := t.Connection.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return ClosedError{err.Error()}
		}
		return ConnectionError{err.Error()}
	}

	return nil
}

// Receive is used to receive a message from the transport. It takes a context
// and blocks until the next message is received from the transport.
func (t *Transport) Receive() (string, []byte, error) {
	_, p, err := t.Connection.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return "", nil, ClosedError{err.Error()}
		}
		return "", nil, ConnectionError{err.Error()}
	}

	msgType, payload, err := Decode(p)
	if err != nil {
		return "", nil, err
	}

	return msgType, payload, nil
}

// Close will cleanly shutdown a websocket connection.
func (t *Transport) Close() error {
	return t.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "bye"))
}
