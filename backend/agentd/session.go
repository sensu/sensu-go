package agentd

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sensu/sensu-go/agent"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/metrics"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/transport"
	"github.com/sirupsen/logrus"
)

const (
	deletedEventSentinel = -1

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second

	// Name of the sessions counter metric
	sessionCounterName = "sensu_go_agent_sessions"

	// Name of the session errors counter metric
	sessionErrorCounterName = "sensu_go_session_errors"

	// Name of the websocket errors metric
	websocketErrorCounterName = "sensu_go_websocket_errors"

	// EventBytesSummaryName is the name of the prometheus summary vec used to
	// track event sizes (in bytes).
	EventBytesSummaryName = "sensu_go_agentd_event_bytes"

	// EventBytesSummaryHelp is the help message for EventBytesSummary
	// Prometheus metrics.
	EventBytesSummaryHelp = "Distribution of event sizes, in bytes, received by agentd on this backend"
)

var (
	eventBytesSummary = metrics.NewEventBytesSummaryVec(EventBytesSummaryName, EventBytesSummaryHelp)

	sessionCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: sessionCounterName,
			Help: "Number of active agent sessions on this backend",
		},
		[]string{"namespace"},
	)
	websocketErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: websocketErrorCounterName,
			Help: "The total number of websocket errors",
		},
		[]string{"op", "error"},
	)
	sessionErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: sessionErrorCounterName,
			Help: "The total number of session errors",
		},
		[]string{"error"},
	)
)

// A Session is a server-side connection between a Sensu backend server and
// the Sensu agent process via the Sensu transport. It is responsible for
// relaying messages to the message bus on behalf of the agent and from the
// bus to the agent from other daemons. It handles transport handshaking and
// transport channel multiplexing/demultiplexing.
type Session struct {
	cfg              SessionConfig
	conn             transport.Transport
	store            store.EntityStore
	storev2          storev2.Interface
	handler          *handler.MessageHandler
	wg               *sync.WaitGroup
	stopWG           sync.WaitGroup
	checkChannel     chan interface{}
	bus              messaging.MessageBus
	ringPool         *ringv2.RingPool
	ctx              context.Context
	cancel           context.CancelFunc
	marshal          agent.MarshalFunc
	unmarshal        agent.UnmarshalFunc
	entityConfig     *entityConfig
	mu               sync.Mutex
	subscriptionsMap map[string]subscription
}

// subscription is used to abstract a message.Subscription and therefore allow
// easier testing
type subscription interface {
	Cancel() error
}

// entityConfig is used by a session to subscribe to entity config updates
type entityConfig struct {
	subscriptions  chan messaging.Subscription
	updatesChannel chan interface{}
}

// Receiver returns the channel for incoming entity updates from the entity
// watcher
func (e *entityConfig) Receiver() chan<- interface{} {
	return e.updatesChannel
}

func newSessionHandler(s *Session) *handler.MessageHandler {
	handler := handler.NewMessageHandler()
	handler.AddHandler(transport.MessageTypeKeepalive, s.handleKeepalive)
	handler.AddHandler(transport.MessageTypeEvent, s.handleEvent)

	return handler
}

// A SessionConfig contains all of the necessary information to initialize
// an agent session.
type SessionConfig struct {
	ContentType   string
	Namespace     string
	AgentAddr     string
	AgentName     string
	User          string
	Subscriptions []string
	WriteTimeout  int

	Bus      messaging.MessageBus
	Conn     transport.Transport
	RingPool *ringv2.RingPool
	Store    store.Store
	Storev2  storev2.Interface

	Marshal   agent.MarshalFunc
	Unmarshal agent.UnmarshalFunc

	// BurialReceiver is used to detect when the previous switch associated
	// with the session has been buried. Necessary when running parallel keepalived
	// workers.
	BurialReceiver *BurialReceiver
}

type BurialReceiver struct {
	ch chan interface{}
}

func NewBurialReceiver() *BurialReceiver {
	return &BurialReceiver{
		ch: make(chan interface{}, 1),
	}
}

func (b *BurialReceiver) Receiver() chan<- interface{} {
	return b.ch
}

// NewSession creates a new Session object given the triple of a transport
// connection, message bus, and store.
// The Session is responsible for stopping itself, and does so when it
// encounters a receive error.
func NewSession(ctx context.Context, cfg SessionConfig) (*Session, error) {
	logger.WithFields(logrus.Fields{
		"addr":          cfg.AgentAddr,
		"namespace":     cfg.Namespace,
		"agent":         cfg.AgentName,
		"subscriptions": cfg.Subscriptions,
	}).Info("agent connected")

	ctx, cancel := context.WithCancel(ctx)

	s := &Session{
		conn:             cfg.Conn,
		cfg:              cfg,
		wg:               &sync.WaitGroup{},
		checkChannel:     make(chan interface{}, 100),
		store:            cfg.Store,
		storev2:          cfg.Storev2,
		bus:              cfg.Bus,
		subscriptionsMap: map[string]subscription{},
		ctx:              ctx,
		cancel:           cancel,
		ringPool:         cfg.RingPool,
		unmarshal:        cfg.Unmarshal,
		marshal:          cfg.Marshal,
		entityConfig: &entityConfig{
			subscriptions:  make(chan messaging.Subscription, 1),
			updatesChannel: make(chan interface{}, 10),
		},
	}

	// Optionally subscribe to burial notifications
	if cfg.BurialReceiver != nil {
		topic := messaging.BurialTopic(cfg.Namespace, cfg.AgentName)
		subscription, err := s.bus.Subscribe(topic, cfg.AgentName, cfg.BurialReceiver)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = subscription.Cancel()
		}()
	}

	if err := s.bus.Publish(messaging.TopicKeepalive, makeEntitySwitchBurialEvent(cfg)); err != nil {
		return nil, err
	}

	// wait for indication that the burial has been processed
	if cfg.BurialReceiver != nil {
		select {
		case <-cfg.BurialReceiver.ch:
		case <-time.After(time.Minute):
			return nil, errors.New("session could not be established after 60s, giving up")
		}
	}

	s.handler = newSessionHandler(s)
	return s, nil
}

// When the session is created, it will send this event to keepalived, ensuring
// that any previously existing switch is buried. This is necessary to make
// sure that the switch is properly recreated if the timeouts have changed.
//
// Keepalived checks for deletedEventSentinel, so that other components can
// message to it that a particular entity's switch can be buried.
func makeEntitySwitchBurialEvent(cfg SessionConfig) *corev2.Event {
	return &corev2.Event{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: cfg.Namespace,
		},
		Entity: &corev2.Entity{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace,
				Name:      cfg.AgentName,
			},
			Subscriptions: cfg.Subscriptions,
			EntityClass:   corev2.EntityAgentClass,
		},
		Check: &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace,
				Name:      corev2.KeepaliveCheckName,
			},
		},
		Timestamp: deletedEventSentinel,
	}
}

// Receiver returns the check channel for the session.
func (s *Session) Receiver() chan<- interface{} {
	return s.checkChannel
}

func (s *Session) receiver() {
	defer func() {
		s.cancel()
		s.wg.Done()
		logger.Info("shutting down agent session: stopping receiver")
	}()

	for {
		if err := s.ctx.Err(); err != nil {
			sessionErrorCounter.WithLabelValues(err.Error()).Inc()
			return
		}
		msg, err := s.conn.Receive()
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError:
				websocketErrorCounter.WithLabelValues("recv", "ConnectionError").Inc()
				logger.WithFields(logrus.Fields{
					"addr":  s.cfg.AgentAddr,
					"agent": s.cfg.AgentName,
				}).WithError(err).Warn("stopping session")
			case transport.ClosedError:
				websocketErrorCounter.WithLabelValues("recv", "ClosedError").Inc()
				logger.WithFields(logrus.Fields{
					"addr":  s.cfg.AgentAddr,
					"agent": s.cfg.AgentName,
				}).WithError(err).Warn("stopping session")
			default:
				websocketErrorCounter.WithLabelValues("recv", "UnknownError").Inc()
				logger.WithError(err).Error("recv error")
			}
			return
		}
		ctx, cancel := context.WithTimeout(s.ctx, time.Duration(s.cfg.WriteTimeout)*time.Second)
		if err := s.handler.Handle(ctx, msg.Type, msg.Payload); err != nil {
			logger.WithError(err).WithFields(logrus.Fields{
				"type":    msg.Type,
				"payload": string(msg.Payload)}).Error("error handling message")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error - boot the agent out of the session
				sessionErrorCounter.WithLabelValues("store.ErrInternal").Inc()
				logger.Error("internal error - stopping session")
				go s.Stop()
			}
		}
		cancel()
	}
}

func (s *Session) sender() {
	defer func() {
		s.cancel()
		s.wg.Done()
		logger.Info("shutting down agent session: stopping sender")
	}()

	for {
		var msg *transport.Message
		select {
		case e := <-s.entityConfig.updatesChannel:
			watchEvent, ok := e.(*store.WatchEventEntityConfig)
			if !ok {
				logger.Errorf("session received unexpected struct: %T", e)
				continue
			}

			// Handle the delete and unknown watch events
			switch watchEvent.Action {
			case store.WatchDelete:
				// stop session
				return
			case store.WatchUnknown:
				logger.Error("session received unknown watch event")
				continue
			}

			if watchEvent.Entity == nil {
				logger.Error("session received nil entity in watch event")
				continue
			}

			lager := logger.WithFields(logrus.Fields{
				"action":    watchEvent.Action.String(),
				"entity":    watchEvent.Entity.Metadata.Name,
				"namespace": watchEvent.Entity.Metadata.Namespace,
			})
			lager.Debug("entity update received")

			// Enforce the entity class to agent
			if watchEvent.Entity.EntityClass != corev2.EntityAgentClass {
				watchEvent.Entity.EntityClass = corev2.EntityAgentClass
				lager.Warningf(
					"misconfigured entity class %q, updating entity to be a %s",
					watchEvent.Entity.EntityClass,
					corev2.EntityAgentClass,
				)

				// Update the entity in the store
				configReq := storev2.NewResourceRequestFromResource(s.ctx, watchEvent.Entity)
				wrapper, err := storev2.WrapResource(watchEvent.Entity)
				if err != nil {
					lager.WithError(err).Error("could not wrap the entity config")
					continue
				}

				if err := s.storev2.CreateOrUpdate(configReq, wrapper); err != nil {
					sessionErrorCounter.WithLabelValues(err.Error()).Inc()
					lager.WithError(err).Error("could not update the entity config")
				}

				// We will not immediately send an update to the agent, but rather wait
				// for the watch event for that entity config
				continue
			}

			bytes, err := s.marshal(watchEvent.Entity)
			if err != nil {
				lager.WithError(err).Error("session failed to serialize entity config")
				continue
			}

			// Determine if some subscriptions were added and/or removed, by first
			// sorting the subscriptions and then comparing those
			s.mu.Lock()
			oldSubscriptions := sortSubscriptions(s.cfg.Subscriptions)
			newSubscriptions := sortSubscriptions(watchEvent.Entity.Subscriptions)
			added, removed := diff(oldSubscriptions, newSubscriptions)
			s.cfg.Subscriptions = newSubscriptions
			s.mu.Unlock()
			if len(added) > 0 {
				lager.Debugf("found %d new subscription(s): %v", len(added), added)
				// The error will already be logged so we can ignore it, and we still
				// want to send the entity config update to the agent
				_ = s.subscribe(added)
			}
			if len(removed) > 0 {
				lager.Debugf("found %d subscription(s) to unsubscribe from: %v", len(removed), removed)
				s.unsubscribe(removed)
			}

			if watchEvent.Entity.Metadata.Labels[corev2.ManagedByLabel] == "sensu-agent" {
				lager.Debug("not sending entity update because entity is managed by its agent")
			}

			msg = transport.NewMessage(transport.MessageTypeEntityConfig, bytes)
		case c := <-s.checkChannel:
			request, ok := c.(*corev2.CheckRequest)
			if !ok {
				logger.Error("session received non-config over check channel")
				continue
			}

			configBytes, err := s.marshal(request)
			if err != nil {
				logger.WithError(err).Error("session failed to serialize check request")
				continue
			}

			msg = transport.NewMessage(corev2.CheckRequestType, configBytes)
		case <-s.ctx.Done():
			return
		}
		logger.WithFields(logrus.Fields{
			"type":         msg.Type,
			"payload_size": len(msg.Payload),
		}).Debug("session - sending message")
		if err := s.conn.Send(msg); err != nil {
			switch err := err.(type) {
			case transport.ConnectionError:
				websocketErrorCounter.WithLabelValues("send", "ConnectionError").Inc()
			case transport.ClosedError:
				websocketErrorCounter.WithLabelValues("send", "ClosedError").Inc()
			default:
				websocketErrorCounter.WithLabelValues("send", "UnknownError").Inc()
				logger.WithError(err).Error("send error")
			}
			return
		}
	}
}

// Start a Session.
// 1. Start sender
// 2. Start receiver
// 3. Start goroutine that waits for context cancellation, and shuts down service.
func (s *Session) Start() (err error) {
	defer close(s.entityConfig.subscriptions)
	sessionCounter.WithLabelValues(s.cfg.Namespace).Inc()
	s.wg = &sync.WaitGroup{}
	s.wg.Add(2)
	s.stopWG.Add(1)
	go s.sender()
	go s.receiver()
	go func() {
		<-s.ctx.Done()
		s.stop()
	}()

	defer func() {
		if err != nil {
			sessionErrorCounter.WithLabelValues("ErrStart").Inc()
			s.cancel()
		}
	}()

	lager := logger.WithFields(logrus.Fields{
		"agent":     s.cfg.AgentName,
		"namespace": s.cfg.Namespace,
	})

	// Subscribe the agent to its entity_config topic
	topic := messaging.EntityConfigTopic(s.cfg.Namespace, s.cfg.AgentName)
	lager.WithField("topic", topic).Debug("subscribing to topic")
	// Get a unique name for the agent, which will be used as the consumer of the
	// bus, in order to avoid problems with an agent reconnecting before its
	// session is ended
	agentName := agentUUID(s.cfg.Namespace, s.cfg.AgentName)
	subscription, err := s.bus.Subscribe(topic, agentName, s.entityConfig)
	if err != nil {
		lager.WithError(err).Error("error starting subscription")
		return err
	}
	s.entityConfig.subscriptions <- subscription

	// Determine if the entity already exists
	req := storev2.NewResourceRequest(s.ctx, s.cfg.Namespace, s.cfg.AgentName, (&corev3.EntityConfig{}).StoreName())
	wrapper, err := s.storev2.Get(req)
	if err != nil {
		// We do not want to send an error if the entity config does not exist
		if _, ok := err.(*store.ErrNotFound); !ok {
			lager.WithError(err).Error("error querying the entity config")
			return err
		}
		lager.Debug("no entity config found")

		// Indicate to the agent that this entity does not exist
		meta := corev2.NewObjectMeta(corev3.EntityNotFound, s.cfg.Namespace)
		watchEvent := &store.WatchEventEntityConfig{
			Action: store.WatchCreate,
			Entity: &corev3.EntityConfig{
				Metadata:    &meta,
				EntityClass: corev2.EntityAgentClass,
			},
		}
		err = s.bus.Publish(messaging.EntityConfigTopic(s.cfg.Namespace, s.cfg.AgentName), watchEvent)
		if err != nil {
			lager.WithError(err).Error("error publishing entity config")
			return err
		}
	} else {
		// An entity config already exists, therefore we should use the stored
		// entity subscriptions rather than what the agent provided us for the
		// subscriptions
		lager.Debug("an entity config was found")

		var storedEntityConfig corev3.EntityConfig
		err = wrapper.UnwrapInto(&storedEntityConfig)
		if err != nil {
			lager.WithError(err).Error("error unwrapping entity config")
			return err
		}

		// Remove the managed_by label if the value is sensu-agent, in case the
		// entity is no longer managed by its agent
		if storedEntityConfig.Metadata.Labels[corev2.ManagedByLabel] == "sensu-agent" {
			delete(storedEntityConfig.Metadata.Labels, corev2.ManagedByLabel)
		}

		// Send back this entity config to the agent so it uses that rather than
		// its local config for its events
		watchEvent := &store.WatchEventEntityConfig{
			Action: store.WatchUpdate,
			Entity: &storedEntityConfig,
		}
		err = s.bus.Publish(messaging.EntityConfigTopic(s.cfg.Namespace, s.cfg.AgentName), watchEvent)
		if err != nil {
			lager.WithError(err).Error("error publishing entity config")
			return err
		}

		// Update the session subscriptions so it uses the stored subscriptions
		s.mu.Lock()
		s.cfg.Subscriptions = storedEntityConfig.Subscriptions
		s.mu.Unlock()
	}

	s.mu.Lock()
	subs := make([]string, len(s.cfg.Subscriptions))
	copy(subs, s.cfg.Subscriptions)
	s.mu.Unlock()

	// Subscribe the session to every configured check subscriptions
	if err := s.subscribe(subs); err != nil {
		return err
	}

	return nil
}

// Stop a running session. This will cause the send and receive loops to
// shutdown. Blocks until the session has shutdown.
func (s *Session) Stop() {
	s.cancel()
	s.wg.Wait()
	s.stopWG.Wait()
}

func (s *Session) stop() {
	defer s.stopWG.Done()
	defer func() {
		if err := s.conn.Close(); err != nil {
			websocketErrorCounter.WithLabelValues("close", "CloseSession").Inc()
			logger.WithError(err).Error("error closing session")
		}
	}()

	sessionCounter.WithLabelValues(s.cfg.Namespace).Dec()

	// Gracefully wait for the send and receiver to exit, but force the websocket
	// connection to close itself after the grace period
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(closeGracePeriod):
		sessionErrorCounter.WithLabelValues("GracePeriodExpired").Inc()
	}

	close(s.entityConfig.updatesChannel)
	close(s.checkChannel)

	// Remove the entity config subscriptions
	for sub := range s.entityConfig.subscriptions {
		if err := sub.Cancel(); err != nil {
			logger.WithError(err).Error("unable to unsubscribe from message bus")
		}
	}

	// Unsubscribe the session from every configured check subscriptions
	s.unsubscribe(s.cfg.Subscriptions)
}

// handleKeepalive is the keepalive message handler.
func (s *Session) handleKeepalive(_ context.Context, payload []byte) error {
	keepalive := &corev2.Event{}
	err := s.unmarshal(payload, keepalive)
	if err != nil {
		return err
	}

	if err := keepalive.Validate(); err != nil {
		return err
	}

	// Not done by event.Validate()
	if keepalive.Timestamp == 0 {
		return errors.New("keepalive contains invalid timestamp")
	}

	keepalive.Entity.Subscriptions = corev2.AddEntitySubscription(keepalive.Entity.Name, keepalive.Entity.Subscriptions)

	return s.bus.Publish(messaging.TopicKeepalive, keepalive)
}

// handleEvent is the event message handler.
func (s *Session) handleEvent(_ context.Context, payload []byte) error {
	// Decode the payload to an event
	event := &corev2.Event{}
	if err := s.unmarshal(payload, event); err != nil {
		return err
	}

	// Validate the received event
	if err := event.Validate(); err != nil {
		return err
	}

	// Add the entity subscription to the subscriptions of this entity
	event.Entity.Subscriptions = corev2.AddEntitySubscription(event.Entity.Name, event.Entity.Subscriptions)

	if event.HasCheck() {
		if event.HasMetrics() {
			eventBytesSummary.WithLabelValues(metrics.EventTypeLabelCheckAndMetrics).Observe(float64(len(payload)))
		} else {
			eventBytesSummary.WithLabelValues(metrics.EventTypeLabelCheck).Observe(float64(len(payload)))
		}
		if event.Check.Name == corev2.KeepaliveCheckName {
			return s.bus.Publish(messaging.TopicKeepaliveRaw, event)
		}
	} else if event.HasMetrics() {
		eventBytesSummary.WithLabelValues(metrics.EventTypeLabelMetrics).Observe(float64(len(payload)))
	}

	return s.bus.Publish(messaging.TopicEventRaw, event)
}

// subscribe adds a subscription to the session for every check subscriptions
// provided
func (s *Session) subscribe(subscriptions []string) error {
	// Prevent any modification to the subscriptions
	s.mu.Lock()
	defer s.mu.Unlock()

	lager := logger.WithFields(logrus.Fields{
		"agent":     s.cfg.AgentName,
		"namespace": s.cfg.Namespace,
	})

	// Get a unique name for the agent, which will be used as the consumer of the
	// bus, in order to avoid problems with an reconnecting before its session is
	// ended
	agent := agentUUID(s.cfg.Namespace, s.cfg.AgentName)

	for _, sub := range subscriptions {
		// Ignore empty subscriptions
		if sub == "" {
			continue
		}

		topic := messaging.SubscriptionTopic(s.cfg.Namespace, sub)

		// Ignore the subscription if the session is already subscribed to it
		if _, ok := s.subscriptionsMap[topic]; ok {
			lager.Debugf("ignoring subscription %q because session is already subscribed", sub)
			continue
		}

		lager.Debugf("subscribing to %q", sub)
		subscription, err := s.bus.Subscribe(topic, agent, s)
		if err != nil {
			lager.WithError(err).Errorf("could not subscribe to %q", sub)
			return err
		}
		s.subscriptionsMap[topic] = &subscription
	}

	return nil
}

// unsubscribe removes a session subscription for every check subscriptions
// provided
func (s *Session) unsubscribe(subscriptions []string) {
	// Prevent any modification to the configured subscriptions and the
	// subscriptions map
	s.mu.Lock()
	defer s.mu.Unlock()

	lager := logger.WithFields(logrus.Fields{
		"agent":     s.cfg.AgentName,
		"namespace": s.cfg.Namespace,
	})

	for _, subscriptionName := range subscriptions {
		topic := messaging.SubscriptionTopic(s.cfg.Namespace, subscriptionName)
		if subscription, ok := s.subscriptionsMap[topic]; ok {
			if err := subscription.Cancel(); err != nil {
				lager.WithError(err).Errorf("unable to unsubscribe from %q", subscriptionName)
				continue
			}

			lager.Debugf("successfully unsubscribed from %q", subscriptionName)

			// Once the subscription is successfully canceled, remove it from our
			// subscriptions map
			delete(s.subscriptionsMap, topic)
		} else {
			lager.Errorf("session was not subscribed to %q", subscriptionName)
		}
	}

	if s.ringPool == nil {
		// This is a bit of a hack - allow ringPool to be nil for the benefit
		// of the tests.
		return
	}

	// Remove the ring for every subscription
	var ringWG sync.WaitGroup
	for _, sub := range subscriptions {
		if strings.HasPrefix(sub, "entity:") {
			// Entity subscriptions don't get rings
			continue
		}
		ringWG.Add(1)
		go func(sub string) {
			defer ringWG.Done()
			ring := s.ringPool.Get(ringv2.Path(s.cfg.Namespace, sub))
			lager.Infof("removing agent from ring for subscription %q", sub)
			if err := ring.Remove(context.Background(), s.cfg.AgentName); err != nil {
				sessionErrorCounter.WithLabelValues("ring.Remove").Inc()
				lager.WithError(err).Error("unable to remove agent from ring")
			}
		}(sub)
	}
	ringWG.Wait()
}

func agentUUID(namespace, name string) string {
	return fmt.Sprintf("%s:%s-%s", namespace, name, uuid.New().String())
}

// diff compares the two given slices and returns the elements that were both
// added and removed in the new slice, in comparison to the old slice. It relies
// on both slices being sorted to properly work.
func diff(old, new []string) ([]string, []string) {
	var added, removed []string
	i, j := 0, 0

	for i < len(old) && j < len(new) {
		c := strings.Compare(old[i], new[j])
		if c == 0 {
			i++
			j++
		} else if c < 0 {
			removed = append(removed, old[i])
			i++
		} else {
			added = append(added, new[j])
			j++
		}
	}

	removed = append(removed, old[i:]...)
	added = append(added, new[j:]...)
	return added, removed
}

func removeEmptySubscriptions(subscriptions []string) []string {
	var s []string
	for _, subscription := range subscriptions {
		if subscription != "" {
			s = append(s, subscription)
		}
	}
	return s
}

func sortSubscriptions(subscriptions []string) []string {
	// Remove empty subscriptions
	subscriptions = removeEmptySubscriptions(subscriptions)

	if sort.StringsAreSorted(subscriptions) {
		return subscriptions
	}

	sortedSubscriptions := append(subscriptions[:0:0], subscriptions...)
	sort.Strings(sortedSubscriptions)
	return sortedSubscriptions
}
