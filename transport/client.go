package transport

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

// Connect causes the transport Client to connect to a given websocket backend.
// This is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func Connect(wsServerURL string, tlsCfg *tls.Config) (Transport, error) {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.DefaultDialer

	if tlsCfg != nil {
		dialer.TLSClientConfig = tlsCfg
	}

	conn, resp, err := dialer.Dial(u.String(), nil)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			return nil, fmt.Errorf("handshake failed with status %d", resp.StatusCode)
		}
		return nil, err
	}

	return NewTransport(conn), nil
}
