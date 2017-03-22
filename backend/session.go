package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

// A Session ...
type Session struct {
	conn    *transport.Transport
	handler *handler.MessageHandler
}

// NewSession ...
func NewSession(conn *transport.Transport) *Session {
	return &Session{
		conn:    conn,
		handler: handler.NewMessageHandler(),
	}
}

// Start a Session
func (s *Session) Start() error {
	handshake := &types.BackendHandshake{}
	hsBytes, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("error marshaling handshake: %s", err.Error())
	}

	// shoot first, ask questions later.
	err = s.conn.Send(types.BackendHandshakeType, hsBytes)
	if err != nil {
		return fmt.Errorf("error sending backend handshake: %s", err.Error())
	}

	t, m, err := s.conn.Receive()
	if err != nil {
		return fmt.Errorf("error receiving agent handshake: %s", err.Error())
	}
	if t != types.AgentHandshakeType {
		return errors.New("no handshake from agent")
	}
	agentHandshake := types.AgentHandshake{}
	if err := json.Unmarshal(m, &agentHandshake); err != nil {
		return fmt.Errorf("error unmarshaling agent handshake: %s", err.Error())
	}

	log.Println("agent connected")

	// setup subscriptions for the client and start a sendpump
	// start readpump
	return nil
}
