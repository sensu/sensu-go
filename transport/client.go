package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/types"
)

// Connect causes the transport Client to connect to a given websocket backend.
// This is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func Connect(wsServerURL string, tlsOpts *types.TLSOptions, requestHeader http.Header) (Transport, error) {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.DefaultDialer

	if tlsOpts != nil {
		dialer.TLSClientConfig, err = tlsOpts.ToTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	conn, resp, err := dialer.Dial(u.String(), requestHeader)
	if err != nil {
		if resp != nil {
			if err == websocket.ErrBadHandshake {
				return nil, fmt.Errorf("handshake failed with status %d", resp.StatusCode)
			}
			return nil, fmt.Errorf("connection failed with status %d", resp.StatusCode)
		}
		return nil, err
	}

	return NewTransport(conn), nil
}
