package transport

import (
	"context"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// A Client provides a reusable, buffering transport to a Sensu backend.
type Client struct {
	wsconn    *websocket.Conn
	readLock  *sync.Mutex
	writeLock *sync.Mutex
}

// Connect causes the transport Client to connect to a given websocket backend.
// This is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func (c *Client) Connect(wsServerURL string) error {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	// Investigate what behavior conn.SetReadDeadline() masks/causes. What causes
	// a read timeout?

	c.wsconn = conn
	c.readLock = &sync.Mutex{}
	c.writeLock = &sync.Mutex{}
	return nil
}

// Send is used to send a message over the transport. It takes a message type
// hint and a serialized payload. Send will block until the message has been
// sent.
func (c *Client) Send(ctx context.Context, msgType string, payload []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	msg := Encode(msgType, payload)
	err := c.wsconn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Receive(ctx context.Context) (string, []byte, error) {
	c.readLock.Lock()
	defer c.readLock.Unlock()

	_, p, err := c.wsconn.ReadMessage()
	if err != nil {
		return "", nil, err
	}

	msgType, payload, err := Decode(p)
	if err != nil {
		return "", nil, err
	}

	return msgType, payload, nil
}
