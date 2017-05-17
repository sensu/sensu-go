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
	space = []byte(" ")
	sep   = []byte("\n")

	pongWait   = pingPeriod + (pingPeriod / 6)
	pingPeriod = 60 * time.Second

	writeWait   = 10 * time.Second
	recvTimeout = 10 * time.Second
	sendTimeout = 10 * time.Second

	sendBufferLength = 20
	recvBufferLength = 20
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
	buf := []byte(msgType)
	buf = append(buf, space...)
	buf = append(buf, payload...)
	buf = append(buf, sep...)
	return buf
}

// Decode a message received from a websocket channel.
func Decode(payload []byte) (string, []byte, error) {
	nl := bytes.Index(payload, space)
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

// Validate if the message is a valid Sensu Transport message.
func (m *Message) Validate() error {
	if m.Type == "" {
		return errors.New("message type cannot be empty")
	}

	return nil
}

// A Conn is a connection between sensu Agents and Backends.
type Conn struct {
	ws        *websocket.Conn
	closeChan chan struct{}
	sendq     chan []byte
	recvq     chan *Message
	errChan   chan error
	closed    bool
	mutex     *sync.Mutex
	err       error
}

// NewConnection creates an initialized Transport and return its pointer.
func NewConnection(conn *websocket.Conn) *Conn {
	c := &Conn{
		ws:        conn,
		closeChan: make(chan struct{}, 1),
		sendq:     make(chan []byte, sendBufferLength),
		recvq:     make(chan *Message, recvBufferLength),
		errChan:   make(chan error, 1),
		closed:    false,
		mutex:     &sync.Mutex{},
		err:       nil,
	}
	go c.readPump()
	go c.writePump()
	return c
}

func (c *Conn) close() {
	fmt.Println("entering close()")
	c.mutex.Lock()
	if c.closed {
		fmt.Println("returning from close()")
		return
	}
	defer c.mutex.Unlock()

	c.error(&ClosedError{})
	c.closed = true
	c.ws.Close()
	// we don't close the sendq to avoid a panic. callers should monitor the
	// connection for a terminal error, and shutdown, allowing the sendq channel
	// to be garbage collected.
	close(c.closeChan)
	close(c.recvq)
}

// most of a *Transport's workflow comes from the websocket chat server example.
// https://github.com/gorilla/websocket/blob/a68708917c6a4f06314ab4e52493cc61359c9d42/examples/chat/conn.go
func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case <-c.closeChan:
			c.ws.WriteMessage(websocket.CloseMessage, []byte{})
			c.error(&ClosedError{})
			return
		case msg := <-c.sendq:
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			writer, err := c.ws.NextWriter(websocket.BinaryMessage)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					c.error(&ConnectionError{})
				} else {
					c.error(&ClosedError{})
				}
				return
			}
			writer.Write(msg)

			// Flush messages from the sendq -- this is only a snapshot, since
			// messages can come in while we're "flushing".
			n := len(c.sendq)
			for i := 0; i < n; i++ {
				_, err := writer.Write(<-c.sendq)
				if err != nil {
					c.error(&ConnectionError{})
					return
				}
			}

			if err := writer.Close(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					c.error(&ConnectionError{})
				} else {
					c.error(&ClosedError{})
				}
				return
			}
		}
	}
}

func (c *Conn) readPump() {
	defer func() {
		c.close()
	}()

	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			c.error(err)
			return
		}

		// We may be receiving multiple messages. We want to separate those by lines
		// decode them, and then handle them all.
		messages := bytes.Split(message, sep)
		for _, msg := range messages {
			// TODO(greg): do something with this error too.
			msgType, payload, _ := Decode(msg)
			msg := &Message{msgType, payload}
			c.recvq <- msg
		}
	}
}

// Send encodes and queues a message for sending over the underlying websocket
// connection. If the connection has been closed, the call to Send will
// timeout.
func (c *Conn) Send(ctx context.Context, m *Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	payload := Encode(m.Type, m.Payload)

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	select {
	case c.sendq <- payload:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Receive a message from the connection,
func (c *Conn) Receive(ctx context.Context) (*Message, error) {
	ctx, cancel := context.WithTimeout(ctx, recvTimeout)
	defer cancel()

	select {
	case msg := <-c.recvq:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Conn) error(err error) {
	select {
	case c.errChan <- err:
	default:
		return
	}
}

// Error blocks until a terminal error on the connection is received. Once an
// error is returned, it will be returned forever.
func (c *Conn) Error() error {
	c.mutex.Lock()
	if c.err != nil {
		c.mutex.Unlock()
		return c.err
	}
	c.mutex.Unlock()

	fmt.Println("reading from errChan")
	c.err = <-c.errChan
	return c.err
}

// Close will cleanly shutdown a websocket connection.
func (c *Conn) Close() {
	c.close()
}
