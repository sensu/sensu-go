package transport

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	v2 "github.com/sensu/core/v2"
)

var ErrTooManyRequests = errors.New("too many requests")

// connect establish the connection to a given websocket backend and returns it
// along with any error encountered
func connect(wsServerURL string, tlsOpts *v2.TLSOptions, requestHeader http.Header, handshakeTimeout int) (*websocket.Conn, http.Header, error) {
	// TODO(grep): configurable max sendq depth
	u, err := url.Parse(wsServerURL)
	if err != nil {
		return nil, nil, err
	}

	if handshakeTimeout < 1 {
		handshakeTimeout = 15
	}
	dialer := websocket.Dialer{
		HandshakeTimeout:	time.Second * time.Duration(handshakeTimeout),
		Proxy:			http.ProxyFromEnvironment,
	}

	if tlsOpts != nil {
		dialer.TLSClientConfig, err = tlsOpts.ToClientTLSConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	conn, resp, err := dialer.Dial(u.String(), requestHeader)
	if err != nil {
		if resp != nil {
			if err == websocket.ErrBadHandshake {
				err := fmt.Errorf("handshake failed with status %d", resp.StatusCode)
				body, berr := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
				if berr == nil {
					err = fmt.Errorf("%s: %s", err, string(body))
				}
				return nil, resp.Header, err
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				return nil, nil, ErrTooManyRequests
			}
			return nil, resp.Header, fmt.Errorf("connection failed with status %d", resp.StatusCode)
		}
		return nil, nil, err
	}

	return conn, resp.Header, nil
}

// Connect causes the transport Client to connect to a given websocket server.
// Transport is a thin wrapper around a websocket connection that makes the
// connection safe for concurrent use by multiple goroutines.
func Connect(wsServerURL string, tlsOpts *v2.TLSOptions, requestHeader http.Header, handshakeTimeout int) (Transport, http.Header, error) {
	conn, resp, err := connect(wsServerURL, tlsOpts, requestHeader, handshakeTimeout)
	if err != nil {
		return nil, nil, err
	}

	return NewTransport(conn), resp, nil
}
