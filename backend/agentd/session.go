package agentd

import (
	"context"
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

// A Session is a server-side connection between a Sensu backend server and
// the Sensu agent process via the Sensu transport. It is responsible for
// relaying messages to the message bus on behalf of the agent and from the
// bus to the agent from other daemons. It handles transport handshaking and
// transport channel multiplexing/demultiplexing.
type Session struct {
	ID string

	conn          transport.Transport
	store         store.Store
	handler       *handler.MessageHandler
	stopping      chan struct{}
	wg            *sync.WaitGroup
	sendq         chan *transport.Message
	subscriptions []string
	checkChannel  chan interface{}
	bus           messaging.MessageBus
}

func newSessionHandler(s *Session) *handler.MessageHandler {
	handler := handler.NewMessageHandler()
	handler.AddHandler(transport.KeepaliveMessageType, s.handleKeepalive)
	handler.AddHandler(transport.EventMessageType, s.handleEvent)

	return handler
}

// NewSession creates a new Session object given the triple of a transport
// connection, message bus, and store.
func NewSession(conn transport.Transport, bus messaging.MessageBus, store store.Store) (*Session, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	s := &Session{
		ID:           id.String(),
		conn:         conn,
		stopping:     make(chan struct{}, 1),
		wg:           &sync.WaitGroup{},
		sendq:        make(chan *transport.Message, 10),
		checkChannel: make(chan interface{}, 100),

		store: store,
		bus:   bus,
	}
	s.handler = newSessionHandler(s)
	return s, nil
}

// the handshake is asynchronous, but the Transport Receive() method is blocking,
// so we send our backend handshake before waiting on the agent handshake.
// Once we have the agent handshake, which is just an Entity object, we
// subscribe to the corresponding channels on the message bus that match the
// agent's subscriptions.
func (s *Session) handshake() error {
	handshake := &types.BackendHandshake{}
	hsBytes, err := json.Marshal(handshake)
	if err != nil {
		return fmt.Errorf("error marshaling handshake: %s", err.Error())
	}

	// shoot first, ask questions later, because Receive() is blocking and
	// Send() is not.
	msg := &transport.Message{
		Type:    types.BackendHandshakeType,
		Payload: hsBytes,
	}

	// TODO(grep): This will block indefinitely. We need to start managing
	// timeouts in transport.WebsocketTransport.
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

	// Validate the agent organization and environment
	ctx := context.TODO()
	if _, err = s.store.GetEnvironment(ctx, agentHandshake.Organization, agentHandshake.Environment); err != nil {
		return fmt.Errorf("the environment '%s:%s' is invalid", agentHandshake.Organization, agentHandshake.Environment)
	}

	s.subscriptions = agentHandshake.Subscriptions
	for _, sub := range s.subscriptions {
		topic := messaging.SubscriptionTopic(agentHandshake.Organization, agentHandshake.Environment, sub)
		logger.Debugf("Subscribing to topic %s", topic)
		if err := s.bus.Subscribe(topic, s.ID, s.checkChannel); err != nil {
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

func (s *Session) recvPump() {
	defer func() {
		logger.Info("session disconnected - stopping recvPump")
		s.wg.Done()
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
			err := s.handler.Handle(msg.Type, msg.Payload)
			if err != nil {
				logger.Error("error handling message: ", msg)
			}
		case <-s.stopping:
			return
		}
	}
}

func (s *Session) subPump() {
	defer func() {
		s.wg.Done()
		logger.Info("shutting down - stopping subPump")
	}()

	for {
		select {
		case c := <-s.checkChannel:
			request, ok := c.(*types.CheckRequest)
			if !ok {
				logger.Errorf("session received non-config over check channel")
				continue
			}

			configBytes, err := json.Marshal(request)
			if err != nil {
				logger.WithError(err).Error("session failed to serialize check request")
			}

			msg := &transport.Message{
				Type:    types.CheckRequestType,
				Payload: configBytes,
			}
			s.sendq <- msg
		case <-s.stopping:
			return
		}
	}
}

func (s *Session) sendPump() {
	defer func() {
		s.wg.Done()
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
		}
	}
}

// Start a Session.
// 1. Perform the handshake (this blocks)
// 2. Start send pump
// 3. Start receive pump
// 4. Start subscription pump
// 5. Ensure bus unsubscribe when the session shuts down.
func (s *Session) Start() error {
	err := s.handshake()
	if err != nil {
		return err
	}

	s.wg = &sync.WaitGroup{}
	s.wg.Add(3)
	go s.sendPump()
	go s.recvPump()
	go s.subPump()

	return nil
}

// Stop a running session. This will cause the send and receive loops to
// shutdown. Blocks until the session has shutdown.
func (s *Session) Stop() {
	close(s.stopping)
	s.wg.Wait()
	for _, sub := range s.subscriptions {
		s.bus.Unsubscribe(sub, s.ID)
	}
	close(s.checkChannel)
}

func (s *Session) handleKeepalive(payload []byte) error {
	keepalive := &types.Event{}
	err := json.Unmarshal(payload, keepalive)
	if err != nil {
		return err
	}

	// TODO(greg): better entity validation than this garbage.
	if keepalive.Entity == nil {
		return errors.New("keepalive does not contain an entity")
	}

	if keepalive.Timestamp == 0 {
		return errors.New("keepalive contains invalid timestamp")
	}

	return s.bus.Publish(messaging.TopicKeepalive, keepalive)
}

func (s *Session) handleEvent(payload []byte) error {
	event := &types.Event{}
	if err := json.Unmarshal(payload, event); err != nil {
		return err
	}
	return s.bus.Publish(messaging.TopicEventRaw, event)
}
