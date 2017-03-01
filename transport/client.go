package transport

import (
	"net/url"

	"github.com/gorilla/websocket"
)

// Connect causes the transport Client to connect to a given websocket backend.
// This is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func Connect(wsServerURL string) (*Transport, error) {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Investigate what behavior conn.SetReadDeadline() masks/causes. What causes
	// a read timeout?
	return NewTransport(conn), nil
}
