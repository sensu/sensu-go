package transport

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Server ...
type Server struct {
	upgrader *websocket.Upgrader
}

// NewServer is used to initialize a new Server and return a pointer to it.
func NewServer() *Server {
	return &Server{
		upgrader: &websocket.Upgrader{},
	}
}

// Serve is used to initialize a websocket connection and returns an pointer to
// a Transport used to communicate with that client.
func (s *Server) Serve(w http.ResponseWriter, r *http.Request) (Transport, error) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewTransport(conn), err
}
