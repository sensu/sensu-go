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

func (s *Session) handshake() error {
	handshake := &types.BackendHandshake{}
	hsBytes, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("error marshaling handshake: %s", err.Error())
	}

	// shoot first, ask questions later.
	msg := &transport.Message{
		Type:    types.BackendHandshakeType,
		Payload: hsBytes,
	}
	err = s.conn.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending backend handshake: %s", err.Error())
	}

	resp, err := s.conn.Receive()
	if err != nil {
		return fmt.Errorf("error receiving agent handshake: %s", err.Error())
	}
	if resp.Type != types.AgentHandshakeType {
		return errors.New("no handshake from agent")
	}
	agentHandshake := types.AgentHandshake{}
	if err := json.Unmarshal(resp.Payload, &agentHandshake); err != nil {
		return fmt.Errorf("error unmarshaling agent handshake: %s", err.Error())
	}

	return nil
}

// Start a Session
func (s *Session) Start() error {
	err := s.handshake()
	if err != nil {
		return err
	}

	log.Println("agent connected")

	return nil
}
