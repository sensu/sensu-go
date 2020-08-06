package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	sep     = []byte("\n")
	msgPool sync.Pool
)

func init() {
	msgPool.New = func() interface{} {
		return &Message{}
	}
}

const (
	// MessageTypeKeepalive is the message type sent for keepalives--which are just an
	// event without a Check or Metrics section.
	MessageTypeKeepalive = "keepalive"

	// MessageTypeEvent is the message type string for events.
	MessageTypeEvent = "event"

	// MessageTypeEntityConfig is the message type sent for entity config updates
	MessageTypeEntityConfig = "entity_config"

	// MessageTypeClose is used to close the connection.
	MessageTypeClose = "close"

	// HeaderKeyAgentName is the HTTP request header specifying the Agent name
	HeaderKeyAgentName = "Sensu-AgentName"

	// HeaderKeyNamespace is the HTTP request header specifying the Agent Namespace
	HeaderKeyNamespace = "Sensu-Namespace"

	// HeaderKeyUser is the HTTP request header specifying the Agent User
	HeaderKeyUser = "Sensu-User"

	// HeaderKeySubscriptions is the HTTP request header specifying the Agent Subscriptions
	HeaderKeySubscriptions = "Sensu-Subscriptions"
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
	// Type is the type of the message (event, etc)
	Type string

	// Payload is the serialized message.
	Payload []byte

	// SendCallback is a callback that is executed after a Send operation.
	// The error value of Send is passed to the callback.
	SendCallback func(error)
}

// The Transport interface defines the set of methods available to a connection
// between the Sensu backend and agent.
type Transport interface {
	// Close will cleanly shutdown a sensu transport connection.
	Close() error

	// Closed returns true if the underlying connection is closed.
	Closed() bool

	// Heartbeat starts a goroutine that sends ping frames to the backend in order
	// to determine if the backend is still responsive
	Heartbeat(ctx context.Context, interval, timeout int)

	// Receive is used to receive a message from the transport. It takes a context
	// and blocks until the next message is received from the transport.
	Receive() (*Message, error)

	// Send is used to send a message over the transport. It takes a message type
	// hint and a serialized payload. Send will block until the message has been
	// sent. Send is synchronous, returning nil if the write to the underlying
	// socket was successful and an error otherwise.
	Send(*Message) error
}

// A WebSocketTransport is a connection between sensu Agents and Backends over
// WebSocket.
type WebSocketTransport struct {
	Connection  *websocket.Conn
	closed      bool
	mutex       *sync.RWMutex
	readTimeout time.Duration
}

// NewTransport creates an initialized Transport and return its pointer.
func NewTransport(conn *websocket.Conn) Transport {
	return &WebSocketTransport{
		Connection: conn,
		closed:     false,
		mutex:      &sync.RWMutex{},
	}
}

// NewMessage creates a new Message.
func NewMessage(msgType string, payload []byte) *Message {
	msg := msgPool.Get().(*Message)
	msg.Type = msgType
	msg.Payload = payload
	return msg
}

// Close attempts to send a "going away" message over the websocket connection.
// This will cause a Write over the websocket transport, which can cause a
// panic. We rescue potential panics and consider the connection closed,
// returning nil, because the connection _will_ be closed. Hay!
func (t *WebSocketTransport) Close() error {
	t.mutex.Lock()

	defer func() {
		// WriteMessage can annoyingly panic, because the websocket conn isn't safe
		// for concurrent use. Recover here, and unlock the mutex.
		_ = recover()
		t.mutex.Unlock()
	}()

	if t.closed {
		return nil
	}

	t.closed = true
	return t.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "bye"))
}

// Closed returns true if the underlying websocket connection has been closed.
func (t *WebSocketTransport) Closed() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.closed
}

// Heartbeat starts a goroutine that sends ping frames to the backend in order
// to determine if the backend is still responsive
func (t *WebSocketTransport) Heartbeat(ctx context.Context, interval, timeout int) {
	if interval < 1 {
		interval = 30
	}
	if timeout < 1 {
		timeout = 45
	}
	if timeout <= interval {
		logger.Warningf("the heartbeat timeout (%d) must be bigger than the heartbeat interval (%d), increasing the timeout", timeout, interval)
		timeout = (interval * 10) / 6
	}
	t.readTimeout = time.Duration(timeout) * time.Second

	pingTicker := time.NewTicker(time.Duration(interval) * time.Second)
	pingWait := t.readTimeout / 2

	go func() {
		defer pingTicker.Stop()
		for {
			select {
			case <-pingTicker.C:
				logger.Debug("sending ping")
				if err := t.Connection.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pingWait)); err != nil {
					logger.WithError(err).Error("could not send a ping to the backend")
					return
				}
			case <-ctx.Done():
				logger.Debug("websocket connection has been closed, stopping the heartbeat")
				return
			}
		}
	}()

	_ = t.Connection.SetReadDeadline(time.Now().Add(t.readTimeout))
	t.Connection.SetPongHandler(func(string) error {
		logger.Debugf("pong received from the backend, setting the read deadline to %d", time.Now().Add(t.readTimeout).Unix())
		return t.Connection.SetReadDeadline(time.Now().Add(t.readTimeout))
	})
}

// Receive a message over the websocket connection. Like Send, returns either
// a ClosedError or a ConnectionError if unable to receive a message. Receive
// blocks until the connection has a message ready or a timeout is reached.
func (t *WebSocketTransport) Receive() (*Message, error) {
	t.mutex.RLock()
	if t.closed {
		t.mutex.RUnlock()
		return nil, ClosedError{"the websocket connection is no longer open"}
	}
	t.mutex.RUnlock()

	_, p, err := t.Connection.ReadMessage()
	if err != nil {
		t.mutex.Lock()
		t.closed = true
		t.mutex.Unlock()

		defer t.Connection.Close()

		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return nil, ClosedError{err.Error()}
		}
		return nil, ConnectionError{err.Error()}
	}
	t.Connection.SetReadDeadline(time.Now().Add(t.readTimeout))

	msgType, payload, err := Decode(p)
	if err != nil {
		return nil, err
	}

	msg := msgPool.Get().(*Message)
	msg.Type = msgType
	msg.Payload = payload
	return msg, nil
}

// Send a message over the websocket connection. If the connection has been
// closed, returns a ClosedError. Returns a ConnectionError if the websocket
// connection returns an error while sending, but the connection is still open.
func (t *WebSocketTransport) Send(m *Message) (err error) {
	defer msgPool.Put(m)
	defer func() {
		if m.SendCallback != nil {
			m.SendCallback(err)
		}
	}()
	t.mutex.RLock()
	if t.closed {
		t.mutex.RUnlock()
		return ClosedError{"the websocket connection is no longer open"}
	}
	t.mutex.RUnlock()

	if m.Type == MessageTypeClose {
		err := t.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			return nil
		}
		return err
	}

	msg := Encode(m.Type, m.Payload)
	if err := t.Connection.WriteMessage(websocket.BinaryMessage, msg); err != nil {
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
