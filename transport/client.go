package transport

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sensu/sensu-go/types"
)

// connect establish the connection to a given websocket backend and returns it
// along with any error encountered
func connect(wsServerURL string, tlsOpts *types.TLSOptions, requestHeader http.Header, handshakeTimeout int) (*websocket.Conn, error) {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return nil, err
	}

	if handshakeTimeout < 1 {
		handshakeTimeout = 15
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: time.Second * time.Duration(handshakeTimeout),
		Proxy:            http.ProxyFromEnvironment,
	}

	if tlsOpts != nil {
		dialer.TLSClientConfig, err = tlsOpts.ToClientTLSConfig()
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

	return conn, nil
}

// Connect causes the transport Client to connect to a given websocket backend.
// This is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func Connect(wsServerURL string, tlsOpts *types.TLSOptions, requestHeader http.Header, handshakeTimeout int) (Transport, error) {
	conn, err := connect(wsServerURL, tlsOpts, requestHeader, handshakeTimeout)
	if err != nil {
		return nil, err
	}

	return NewTransport(conn), nil
}
