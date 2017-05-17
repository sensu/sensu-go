package agentd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
)

// A Session ...
type Session struct {
	ID string

	conn          *transport.Transport
	store         store.Store
	handler       handler.MessageHandler
	stopping      chan struct{}
	stopped       chan struct{}
	sendq         chan *transport.Message
	subscriptions []string
	checkChannel  chan []byte
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

		store: store,
		bus:   bus,
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

	logger.Infof("agent connected: id=%s subscriptions=%s", agentHandshake.ID, agentHandshake.Subscriptions)

	return nil
}

func (s *Session) receiveMessages(out chan *transport.Message) {
	defer close(out)
	for {
		m, err := s.conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError, transport.ClosedError:
				logger.Error("recv error: ", err.Error())
				return
			default:
				logger.Error("recv error: ", err.Error())
				continue
			}
		}
		out <- m
	}
}

func (s *Session) recvPump(wg *sync.WaitGroup) {
	defer func() {
		logger.Info("session disconnected - stopping recvPump")
		wg.Done()
	}()

	msgChannel := make(chan *transport.Message)
	go s.receiveMessages(msgChannel)
	for {
		select {
		case msg, ok := <-msgChannel:
			if !ok {
				return
			}

			logger.Debugf("session - received message: %s", string(msg.Payload))
			err := s.handler.Handle(msg)
			if err != nil {
				logger.Error("error handling message: ", msg)
			}
		case <-s.stopping:
			return
		}
	}
}

func (s *Session) subPump(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		logger.Info("shutting down - stopping subPump")
	}()

	for {
		select {
		case check := <-s.checkChannel:
			msg := &transport.Message{
				Type:    types.EventType,
				Payload: check,
			}
			s.sendq <- msg
		case <-s.stopping:
			return
		default:
			if s.conn.Closed() {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (s *Session) sendPump(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		logger.Info("shutting down - stopping sendPump")
	}()

	for {
		select {
		case msg := <-s.sendq:
			logger.Debugf("session - sending message: %s", string(msg.Payload))
			err := s.conn.Send(msg)
			if err != nil {
				switch err := err.(type) {
				case transport.ConnectionError, transport.ClosedError:
					return
				default:
					logger.Error("send error:", err.Error())
				}
			}
		case <-s.stopping:
			return
		default:
			if s.conn.Closed() {
				return
			}
			time.Sleep(1 * time.Millisecond)
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

	if keepalive.Timestamp == 0 {
		return errors.New("keepalive contains invalid timestamp")
	}

	return s.bus.Publish(messaging.TopicKeepalive, payload)
}

func (s *Session) handleEvent(payload []byte) error {
	return s.bus.Publish(messaging.TopicEventRaw, payload)
}
