package transport

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = &websocket.Upgrader{}
)

// Serve is used to initialize a websocket connection and returns an pointer to
// a Transport used to communicate with that client.
func Serve(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewConnection(conn), err
}
