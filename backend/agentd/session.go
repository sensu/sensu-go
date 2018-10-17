package agentd

import (
	"context"
	"errors"
	"fmt"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigDefault

// SessionStore specifies the storage requirements of the Session.
type SessionStore interface {
	store.EntityStore
	store.NamespaceStore
}

// A Session is a server-side connection between a Sensu backend server and
// the Sensu agent process via the Sensu transport. It is responsible for
// relaying messages to the message bus on behalf of the agent and from the
// bus to the agent from other daemons. It handles transport handshaking and
// transport channel multiplexing/demultiplexing.
type Session struct {
	cfg          SessionConfig
	conn         transport.Transport
	store        SessionStore
	handler      *handler.MessageHandler
	stopping     chan struct{}
	wg           *sync.WaitGroup
	sendq        chan *transport.Message
	checkChannel chan interface{}
	bus          messaging.MessageBus

	subscriptions chan messaging.Subscription
}

func newSessionHandler(s *Session) *handler.MessageHandler {
	handler := handler.NewMessageHandler()
	handler.AddHandler(transport.MessageTypeKeepalive, s.handleKeepalive)
	handler.AddHandler(transport.MessageTypeEvent, s.handleEvent)

	return handler
}

// A SessionConfig contains all of the ncessary information to initialize
// an agent session.
type SessionConfig struct {
	Namespace     string
	AgentAddr     string
	AgentID       string
	User          string
	Subscriptions []string
}

// NewSession creates a new Session object given the triple of a transport
// connection, message bus, and store.
// The Session is responsible for stopping itself, and does so when it
// encounters a receive error.
func NewSession(cfg SessionConfig, conn transport.Transport, bus messaging.MessageBus, store Store) (*Session, error) {
	// Validate the agent namespace
	ctx := context.TODO()
	if _, err := store.GetNamespace(ctx, cfg.Namespace); err != nil {
		return nil, fmt.Errorf("the namespace '%s' is invalid", cfg.Namespace)
	}

	logger.WithFields(logrus.Fields{
		"addr":          cfg.AgentAddr,
		"id":            cfg.AgentID,
		"subscriptions": cfg.Subscriptions,
	}).Info("agent connected")

	s := &Session{
		conn:          conn,
		cfg:           cfg,
		stopping:      make(chan struct{}, 1),
		wg:            &sync.WaitGroup{},
		sendq:         make(chan *transport.Message, 10),
		checkChannel:  make(chan interface{}, 100),
		store:         store,
		bus:           bus,
		subscriptions: make(chan messaging.Subscription, len(cfg.Subscriptions)),
	}
	s.handler = newSessionHandler(s)
	return s, nil
}

// Receiver returns the check channel for the session.
func (s *Session) Receiver() chan<- interface{} {
	return s.checkChannel
}

func (s *Session) recvPump() {
	defer func() {
		logger.Info("session disconnected - stopping recvPump")
		s.wg.Done()
		select {
		case <-s.stopping:
		default:
			s.Stop()
		}
	}()

	for {
		select {
		case <-s.stopping:
			return
		default:
		}
		msg, err := s.conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError, transport.ClosedError:
				logger.WithFields(logrus.Fields{
					"addr":       s.cfg.AgentAddr,
					"id":         s.cfg.AgentID,
					"recv error": err.Error(),
				}).Warn("stopping session")
				return
			default:
				logger.WithError(err).Error("recv error")
				continue
			}
		}
		if err := s.handler.Handle(msg.Type, msg.Payload); err != nil {
			logger.WithError(err).WithFields(logrus.Fields{
				"type":    msg.Type,
				"payload": string(msg.Payload)}).Error("error handling message")
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
				logger.Error("session received non-config over check channel")
				continue
			}

			configBytes, err := json.Marshal(request)
			if err != nil {
				logger.WithError(err).Error("session failed to serialize check request")
				continue
			}

			msg := transport.NewMessage(types.CheckRequestType, configBytes)
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
			logger.WithField("message", string(msg.Payload)).Debug("session - sending message")
			err := s.conn.Send(msg)
			if err != nil {
				switch err := err.(type) {
				case transport.ConnectionError, transport.ClosedError:
					return
				default:
					logger.WithError(err).Error("send error")
				}
			}
		case <-s.stopping:
			return
		}
	}
}

// Start a Session.
// 1. Start send pump
// 2. Start receive pump
// 3. Start subscription pump
// 5. Ensure bus unsubscribe when the session shuts down.
func (s *Session) Start() (err error) {
	s.wg = &sync.WaitGroup{}
	s.wg.Add(3)
	go s.sendPump()
	go s.recvPump()
	go s.subPump()

	namespace := s.cfg.Namespace
	agentID := fmt.Sprintf("%s:%s", namespace, s.cfg.AgentID)

	defer func() {
		if err != nil {
			s.Stop()
		}
	}()

	for _, sub := range s.cfg.Subscriptions {
		// Ignore empty subscriptions
		if sub == "" {
			continue
		}

		topic := messaging.SubscriptionTopic(namespace, sub)
		logger.WithField("topic", topic).Debug("subscribing to topic")
		subscription, err := s.bus.Subscribe(topic, agentID, s)
		if err != nil {
			logger.WithError(err).Error("error starting subscription")
			return err
		}
		s.subscriptions <- subscription
	}
	close(s.subscriptions)

	return nil
}

// Stop a running session. This will cause the send and receive loops to
// shutdown. Blocks until the session has shutdown.
func (s *Session) Stop() {
	close(s.stopping)
	s.wg.Wait()

	for sub := range s.subscriptions {
		if err := sub.Cancel(); err != nil {
			logger.WithError(err).Error("unable to unsubscribe from message bus")
		}
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

	keepalive.Entity.Subscriptions = addEntitySubscription(keepalive.Entity.ID, keepalive.Entity.Subscriptions)

	return s.bus.Publish(messaging.TopicKeepalive, keepalive)
}

func (s *Session) handleEvent(payload []byte) error {
	// Decode the payload to an event
	event := &types.Event{}
	if err := json.Unmarshal(payload, event); err != nil {
		return err
	}

	// Validate the received event
	if err := event.Validate(); err != nil {
		return err
	}

	// Verify if we have a source in the event and if so, use it as the entity by
	// creating or retrieving it from the store
	if event.HasCheck() {
		if err := getProxyEntity(event, s.store); err != nil {
			return err
		}
	}

	// Add the entity subscription to the subscriptions of this entity
	event.Entity.Subscriptions = addEntitySubscription(event.Entity.ID, event.Entity.Subscriptions)

	return s.bus.Publish(messaging.TopicEventRaw, event)
}
