package agentd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

// A Session ...
type Session struct {
	ID string

	conn          *transport.Transport
	store         store.Store
	handler       *handler.MessageHandler
	stopping      chan struct{}
	stopped       chan struct{}
	sendq         chan *transport.Message
	subscriptions []string
	checkChannel  chan []byte
	disconnected  bool
	bus           messaging.MessageBus
}

func newSessionHandler(s *Session) *handler.MessageHandler {
	handler := handler.NewMessageHandler()
	handler.AddHandler(types.KeepaliveType, s.handleKeepalive)
	handler.AddHandler(types.EventType, s.handleEvent)

	return handler
}

// NewSession ...
func NewSession(conn *transport.Transport, bus messaging.MessageBus, store store.Store) (*Session, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	s := &Session{
		ID:           id.String(),
		conn:         conn,
		stopping:     make(chan struct{}, 1),
		stopped:      make(chan struct{}),
		sendq:        make(chan *transport.Message, 10),
		checkChannel: make(chan []byte, 100),

		disconnected: false,
		store:        store,
		bus:          bus,
	}
	s.handler = newSessionHandler(s)
	return s, nil
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

	s.subscriptions = agentHandshake.Subscriptions
	for _, sub := range s.subscriptions {
		if err := s.bus.Subscribe(sub, s.ID, s.checkChannel); err != nil {
			return err
		}
	}

	logger.Infof("agent connected: id=%s subscriptions=%s\n", agentHandshake.ID, agentHandshake.Subscriptions)

	return nil
}

func (s *Session) recvPump(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if s.disconnected {
			logger.Info("session disconnected - stopping recvPump")
			return
		}

		msg, err := s.conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError:
				s.disconnected = true
			case transport.ClosedError:
				s.disconnected = true
			default:
				logger.Error("recv error:", err.Error())
			}
			continue
		}

		err = s.handler.Handle(msg.Type, msg.Payload)
		if err != nil {
			logger.Error("error handling message: ", msg)
		}
	}
}

func (s *Session) subPump(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if s.disconnected {
			logger.Info("session disconnected - stopping sendPump")
			return
		}

		select {
		case check := <-s.checkChannel:
			msg := &transport.Message{
				Type:    types.EventType,
				Payload: check,
			}
			s.sendq <- msg
		case <-s.stopping:
			logger.Info("shutting down - stopping subPump")
			return
		}

	}
}

func (s *Session) sendPump(wg *sync.WaitGroup) {
	defer wg.Done()

	for {

		select {
		case msg := <-s.sendq:
			err := s.conn.Send(msg)
			if err != nil {
				switch err := err.(type) {
				case transport.ConnectionError:
					s.disconnected = true
				case transport.ClosedError:
					s.disconnected = true
				default:
					logger.Error("send error:", err.Error())
				}
			}
		case <-s.stopping:
			s.disconnected = true
			logger.Info("shutting down - stopping sendPump")
			return
		}
	}
}

// Start a Session
func (s *Session) Start() error {
	err := s.handshake()
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go s.sendPump(wg)
	go s.recvPump(wg)
	go s.subPump(wg)
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		for _, sub := range s.subscriptions {
			s.bus.Unsubscribe(sub, s.ID)
		}
		close(s.checkChannel)
		close(s.stopped)
	}(wg)

	return nil
}

// Stop a running session. This will cause the send and receive loops to
// shutdown. Blocks until the session has shutdown.
func (s *Session) Stop() {
	close(s.stopping)
	<-s.stopped
}

func (s *Session) handleKeepalive(payload []byte) error {
	keepalive := &types.Event{}
	err := json.Unmarshal(payload, keepalive)
	if err != nil {
		return err
	}

	// TODO(greg): better entity validation than this garbaje.
	if keepalive.Entity == nil {
		return errors.New("keepalive does not contain an entity")
	}

	logger.Debug("handling keepalive: ", *keepalive)
	return s.store.UpdateEntity(keepalive.Entity)
}

func (s *Session) handleEvent(payload []byte) error {
	return s.bus.Publish(messaging.TopicEvent, payload)
}
