package keepalived

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sirupsen/logrus"
)

const (
	// KeepaliveCounterVec is the name of the prometheus metric that Sensu
	// exports for counting keepalive events, both dead and alive.
	KeepaliveCounterVec = "sensu_go_keepalives"

	// KeepaliveCounterLabelName represents the status of a counted keepalive.
	KeepaliveCounterLabelName = "status"

	// KeepaliveCounterLabelAlive represents a call to alive().
	KeepaliveCounterLabelAlive = "alive"

	// KeepaliveCounterLabelDead represents a call to dead().
	KeepaliveCounterLabelDead = "dead"
)

var KeepalivesProcessed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: KeepaliveCounterVec,
		Help: "The total name of processed keepalives",
	},
	[]string{KeepaliveCounterLabelName},
)

func init() {
	KeepalivesProcessed.WithLabelValues(KeepaliveCounterLabelAlive)
	KeepalivesProcessed.WithLabelValues(KeepaliveCounterLabelDead)
	_ = prometheus.Register(KeepalivesProcessed)
}

const deletedEventSentinel = -1

// Keepalived is responsible for monitoring keepalive events and recording
// keepalives for entities.
type Keepalived struct {
	bus                     messaging.MessageBus
	workerCount             int
	store                   storev2.Interface
	deregistrationPipelines []*corev2.ResourceReference
	mu                      *sync.Mutex
	wg                      *sync.WaitGroup
	keepaliveChan           chan interface{}
	subscription            messaging.Subscription
	errChan                 chan error
	ctx                     context.Context
	cancel                  context.CancelFunc
	storeTimeout            time.Duration
	reconstructionPeriod    time.Duration
	operatorConcierge       store.OperatorConcierge
	operatorMonitor         store.OperatorMonitor
	backendName             string
}

// Option is a functional option.
type Option func(*Keepalived) error

// Config configures Keepalived.
type Config struct {
	Store                   storev2.Interface
	EventStore              store.EventStore
	Bus                     messaging.MessageBus
	DeregistrationPipelines []*corev2.ResourceReference
	BufferSize              int
	WorkerCount             int
	StoreTimeout            time.Duration
	OperatorConcierge       store.OperatorConcierge
	OperatorMonitor         store.OperatorMonitor
	BackendName             string
}

// New creates a new Keepalived.
func New(c Config, opts ...Option) (*Keepalived, error) {
	if c.BufferSize == 0 {
		logger.Warn("BufferSize not set")
		c.BufferSize = 1
	}
	if c.WorkerCount == 0 {
		logger.Warn("WorkerCount not set")
		c.WorkerCount = 1
	}
	if c.StoreTimeout == 0 {
		logger.Warn("StoreTimeout not set")
		c.StoreTimeout = time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	k := &Keepalived{
		store:                   c.Store,
		bus:                     c.Bus,
		deregistrationPipelines: c.DeregistrationPipelines,
		keepaliveChan:           make(chan interface{}, c.BufferSize),
		workerCount:             c.WorkerCount,
		mu:                      &sync.Mutex{},
		errChan:                 make(chan error, 1),
		ctx:                     ctx,
		cancel:                  cancel,
		storeTimeout:            c.StoreTimeout,
		reconstructionPeriod:    time.Second * 120,
		operatorConcierge:       c.OperatorConcierge,
		operatorMonitor:         c.OperatorMonitor,
		backendName:             c.BackendName,
	}
	for _, o := range opts {
		if err := o(k); err != nil {
			return nil, err
		}
	}
	return k, nil
}

// Receiver returns the keepalive receiver channel.
func (k *Keepalived) Receiver() chan<- interface{} {
	return k.keepaliveChan
}

// Start starts the daemon, returning an error if preconditions for startup
// fail.
func (k *Keepalived) Start() error {
	sub, err := k.bus.Subscribe(messaging.TopicKeepalive, "keepalived", k)
	if err != nil {
		return err
	}

	k.subscription = sub

	k.wg = &sync.WaitGroup{}
	k.startWorkers()
	go k.monitorOperators(k.ctx)

	return nil
}

func (k *Keepalived) monitorOperators(ctx context.Context) {
	req := store.MonitorOperatorsRequest{
		Type:           store.AgentOperator,
		ControllerType: store.BackendOperator,
		ControllerName: k.backendName,
		Every:          10 * time.Second,
		ErrorHandler: func(err error) {
			logger.WithError(err).Error("error monitoring agent keepalives")
		},
	}
	stateCh := k.operatorMonitor.MonitorOperators(ctx, req)
	for {
		select {
		case <-ctx.Done():
			return
		case states := <-stateCh:
			for _, state := range states {
				if err := k.handleNotification(ctx, state); err != nil {
					logger.WithError(err).Error("error handling keepalive notification")
				}
			}
		}
	}
}

// Stop stops the daemon, returning an error if one was encountered during
// shutdown.
func (k *Keepalived) Stop() error {
	k.cancel()
	err := k.subscription.Cancel()
	close(k.keepaliveChan)
	k.wg.Wait()
	close(k.errChan)
	return err
}

// Err returns a channel that the caller can use to listen for terminal errors
// indicating a premature shutdown of the Daemon.
func (k *Keepalived) Err() <-chan error {
	return k.errChan
}

// Name returns the daemon name
func (k *Keepalived) Name() string {
	return "keepalived"
}

func (k *Keepalived) startWorkers() {
	k.wg.Add(k.workerCount)

	for i := 0; i < k.workerCount; i++ {
		go k.processKeepalives(k.ctx)
	}
}

func (k *Keepalived) processKeepalives(ctx context.Context) {
	defer k.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-k.keepaliveChan:
			if !ok {
				return
			}
			event, ok := msg.(*corev2.Event)
			if !ok {
				logger.Error("keepalived received non-Event on keepalive channel")
				continue
			}

			entity := event.Entity
			if entity == nil {
				logger.Error("keepalive channel received keepalive with nil event")
				continue
			}

			if err := entity.Validate(); err != nil {
				logger.WithError(err).Error("invalid keepalive event")
				continue
			}

			if event.Check == nil {
				logger.Error("keepalive event has no check")
				continue
			}

			id := path.Join(entity.Namespace, entity.Name)

			if event.Timestamp == deletedEventSentinel {
				// The keepalive event was deleted, so the concierge should check it out
				tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
				err := k.operatorConcierge.CheckOut(tctx, store.OperatorKey{Namespace: entity.Namespace, Name: entity.Name, Type: store.AgentOperator})
				cancel()
				if err != nil {
					if _, ok := err.(*store.ErrInternal); ok {
						// Fatal error
						select {
						case k.errChan <- err:
						case <-ctx.Done():
						}
						return
					}
					logger.WithError(err).Error("error deleting keepalive")
				}
				// ignore error as this message is advisory
				_ = k.bus.Publish(messaging.BurialTopic(event.Entity.Namespace, event.Entity.Name), nil)
				continue
			}

			if err := k.handleEntityRegistration(entity, event); err != nil {
				logger.WithError(err).Error("error handling entity registration")
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					select {
					case k.errChan <- err:
					case <-ctx.Done():
					}
					return
				}
			}

			warning := int(event.Check.Timeout)
			critical := int(event.Check.Ttl)
			interval := int(event.Check.Interval)
			ttl := time.Duration(warning) * time.Second

			tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
			agentMeta := agentMetadata{
				Warning:  warning,
				Critical: critical,
				Interval: interval,
			}
			metadata, _ := json.Marshal(agentMeta)
			state := store.OperatorState{
				Namespace:      entity.Namespace,
				Name:           entity.Name,
				Type:           store.AgentOperator,
				CheckInTimeout: ttl,
				Present:        true,
				Controller: &store.OperatorKey{
					Name: k.backendName,
					Type: store.BackendOperator,
				},
				Metadata: (*json.RawMessage)(&metadata),
			}
			err := k.operatorConcierge.CheckIn(tctx, state)
			cancel()
			if err != nil {
				logger.WithError(err).Errorf("error checking-in entity %q", id)
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					select {
					case k.errChan <- err:
					case <-ctx.Done():
					}
					return
				}
				continue
			}

			if err := k.handleUpdate(event); err != nil {
				logger.WithError(err).Error("error updating event")
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					select {
					case k.errChan <- err:
					case <-ctx.Done():
					}
					return
				}
			}
		}
	}
}

// HandleError logs an error
func (k *Keepalived) HandleError(err error) {
	logger.WithError(err).Error(err)
}

func (k *Keepalived) handleEntityRegistration(entity *corev2.Entity, event *corev2.Event) error {
	if entity.EntityClass != corev2.EntityAgentClass {
		return nil
	}

	ctx := corev2.SetContextFromResource(k.ctx, entity)
	tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
	defer cancel()

	config, _ := corev3.V2EntityToV3(entity)

	exists := true
	entityConfigStore := storev2.Of[*corev3.EntityConfig](k.store)
	storedEntityConfig, err := entityConfigStore.Get(tctx, storev2.ID{Namespace: entity.Namespace, Name: entity.Name})
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			logger.WithError(err).Error("error while checking if entity exists")
			return err
		}
		exists = false
	}

	if exists {
		// Determine if the entity is managed by its agent
		if entity.ObjectMeta.Labels[corev2.ManagedByLabel] == "sensu-agent" {
			// If this keepalive is the first one sent by an agent, we want to update
			// the stored entity config to reflect the sent one
			if event.Sequence == 1 {
				if err := entityConfigStore.UpdateIfExists(tctx, config); err != nil {
					logger.WithError(err).Error("could not update entity")
					return err
				}
			}
			return nil
		}

		// Determine if this entity was previously managed by its agent but it's no
		// longer the case, in which case we need to reflect that in the stored
		// entity config
		if storedEntityConfig.Metadata.Labels[corev2.ManagedByLabel] == "sensu-agent" {
			// Remove the managed_by label and update the stored entity config
			delete(storedEntityConfig.Metadata.Labels, corev2.ManagedByLabel)
			if err := entityConfigStore.UpdateIfExists(tctx, storedEntityConfig); err != nil {
				logger.WithError(err).Error("could not update entity")
				return err
			}
		}
		return nil
	}

	// The entity config does not exist so create it and publish a registration
	// event
	err = entityConfigStore.CreateIfNotExists(tctx, config)
	if err == nil {
		event := createRegistrationEvent(entity)
		return k.bus.Publish(messaging.TopicEvent, event)
	} else if _, ok := err.(*store.ErrAlreadyExists); ok {
		logger.WithError(err).Warn("received a check event before entity registration")
		return nil
	} else {
		return err
	}
}

func createKeepaliveEvent(rawEvent *corev2.Event) *corev2.Event {
	check := rawEvent.Check
	if check == nil {
		check = &corev2.Check{
			Interval: agent.DefaultKeepaliveInterval,
			Timeout:  corev2.DefaultKeepaliveTimeout,
		}
	}

	// Use the entity keepalive handlers if defined, otherwise fallback to the
	// default keepalive handler
	handlers := []string{corev2.KeepaliveHandlerName}
	if len(rawEvent.Entity.KeepaliveHandlers) > 0 {
		handlers = rawEvent.Entity.KeepaliveHandlers
	}

	// Use the entity keepalive pipelines if defined, otherwise fallback to the
	// default keepalive pipeline
	pipelines := []*corev2.ResourceReference{
		{
			Name:       corev2.KeepalivePipelineName,
			Type:       "Pipeline",
			APIVersion: "core/v2",
		},
	}
	if len(rawEvent.Pipelines) > 0 {
		pipelines = rawEvent.Entity.KeepalivePipelines
	}

	keepaliveCheck := &corev2.Check{
		ObjectMeta: corev2.ObjectMeta{
			Name:      corev2.KeepaliveCheckName,
			Namespace: rawEvent.Entity.Namespace,
		},
		Interval:  check.Interval,
		Timeout:   check.Timeout,
		Ttl:       check.Ttl,
		Handlers:  handlers,
		Pipelines: pipelines,
		Executed:  time.Now().Unix(),
		Issued:    time.Now().Unix(),
		Scheduler: corev2.EtcdScheduler,
	}
	keepaliveEvent := &corev2.Event{
		ObjectMeta: rawEvent.ObjectMeta,
		Timestamp:  time.Now().Unix(),
		Entity:     rawEvent.Entity,
		Check:      keepaliveCheck,
		Pipelines:  pipelines,
	}

	uid, _ := uuid.NewRandom()
	keepaliveEvent.ID = uid[:]
	keepaliveEvent.Sequence = rawEvent.Sequence

	return keepaliveEvent
}

func createRegistrationEvent(entity *corev2.Entity) *corev2.Event {
	registrationCheck := &corev2.Check{
		ObjectMeta: corev2.ObjectMeta{
			Name:      corev2.RegistrationCheckName,
			Namespace: entity.Namespace,
		},
		Interval: 1,
		Handlers: []string{corev2.RegistrationHandlerName},
		Pipelines: []*corev2.ResourceReference{
			{
				Name:       corev2.RegistrationPipelineName,
				Type:       "Pipeline",
				APIVersion: "core/v2",
			},
		},
		Status: 1,
	}
	registrationEvent := &corev2.Event{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: entity.Namespace,
		},
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     registrationCheck,
	}

	return registrationEvent
}

func (k *Keepalived) alive(state store.OperatorState) {
	KeepalivesProcessed.WithLabelValues(KeepaliveCounterLabelAlive).Inc()

	if logrus.GetLevel() == logrus.DebugLevel {
		// avoid unnecessary allocations here
		lager := logger.WithFields(logrus.Fields{
			"present":   true,
			"entity":    state.Name,
			"namespace": state.Namespace,
		})

		lager.Debug("entity is alive")
	}
}

func (k *Keepalived) handleNotification(ctx context.Context, state store.OperatorState) error {
	KeepalivesProcessed.WithLabelValues(KeepaliveCounterLabelDead).Inc()
	lager := logger.WithFields(logrus.Fields{
		"present":       false,
		"entity":        state.Name,
		"operator_type": state.Type.String(),
		"namespace":     state.Namespace,
	})

	lager.Warn("keepalive timed out")

	key := store.OperatorKey{Namespace: state.Namespace, Name: state.Name, Type: store.AgentOperator}

	entityStore := storev2.Of[*corev3.EntityConfig](k.store)
	entityConfig, err := entityStore.Get(ctx, storev2.ID{Namespace: state.Namespace, Name: state.Name})
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			// The entity has been deleted, there is no longer a need to track
			// keepalives for it.
			return k.operatorConcierge.CheckOut(ctx, key)
		}
		return err
	}
	ctx = store.NamespaceContext(ctx, state.Namespace)
	currentEvent, err := k.store.GetEventStore().GetEventByEntityCheck(ctx, state.Name, "keepalive")
	if err != nil {
		lager.WithError(err).Error("error while reading event")
		return err
	}
	if currentEvent == nil {
		// The keepalive was deleted, checkout the operator
		lager.Debug("nil event")
		return k.operatorConcierge.CheckOut(ctx, key)
	}

	if entityConfig.Deregister {
		deregisterer := &Deregistration{
			Store:        k.store,
			MessageBus:   k.bus,
			StoreTimeout: k.storeTimeout,
		}
		if err := deregisterer.Deregister(currentEvent.Entity); err != nil {
			lager.WithError(err).Error("error deregistering entity")
		}
		lager.Debug("deregistering entity")
		return k.operatorConcierge.CheckOut(ctx, key)
	}

	// emit keepalive event on bus
	event := createKeepaliveEvent(currentEvent)
	timeSinceLastSeen := time.Now().Unix() - event.Entity.LastSeen
	warningTimeout := int64(event.Check.Timeout)
	criticalTimeout := event.Check.Ttl
	var timeout int64
	if warningTimeout != 0 && timeSinceLastSeen >= warningTimeout {
		// warning keepalive
		timeout = warningTimeout
		event.Check.Status = 1
	}
	if criticalTimeout != 0 && timeSinceLastSeen >= criticalTimeout {
		// critical keepalive
		timeout = criticalTimeout
		event.Check.Status = 2
	}
	event.Check.Output = fmt.Sprintf("No keepalive sent from %s for %v seconds (>= %v)", event.Entity.Name, timeSinceLastSeen, timeout)

	if err := k.bus.Publish(messaging.TopicEventRaw, event); err != nil {
		lager.WithError(err).Error("error publishing event")
		return err
	}

	var meta agentMetadata
	if err := json.Unmarshal(*state.Metadata, &meta); err != nil {
		lager.WithError(err).Error("error reading state metadata")
		return err
	}

	if interval := time.Duration(meta.Interval) * time.Second; state.CheckInTimeout != interval {
		state.CheckInTimeout = interval
		state.Present = false // defensive, likely unnecessary
		// update state so that its check-in timeout is the keepalive interval
		if err := k.operatorConcierge.CheckIn(ctx, state); err != nil {
			return err
		}
	}

	if event.Entity.EntityClass != corev2.EntityAgentClass {
		// keepalives for non-agent entities won't affect ring things
		return nil
	}

	for _, sub := range event.Entity.Subscriptions {
		// TODO(eric) figure out round robin things
		// lager := lager.WithFields(logrus.Fields{"subscription": sub})
		// if strings.HasPrefix(sub, "entity:") {
		// 	// Entity subscriptions don't get rings
		// 	continue
		// }
		// ring := k.ringPool.Get(ringv2.Path(namespace, sub))
		// if err := ring.Remove(ctx, name); err != nil {
		// 	lager.WithError(err).Error("error removing entity from ring")
		// 	continue
		// }
		lager.Trace("removing entity from ring", sub)
	}

	return nil
}

type agentMetadata struct {
	Warning  int `json:"w"`
	Critical int `json:"c"`
	Interval int `json:"i"`
}

// handleUpdate sets the entity's last seen time and publishes an OK check event
// to the message bus.
func (k *Keepalived) handleUpdate(e *corev2.Event) error {
	if e.Check == nil {
		return errors.New("no check in keepalive event")
	}
	entity := e.Entity

	entity.LastSeen = e.Timestamp
	_, entityState := corev3.V2EntityToV3(entity)

	entityStateStore := storev2.Of[*corev3.EntityState](k.store)

	if err := entityStateStore.CreateOrUpdate(k.ctx, entityState); err != nil {
		logger.WithError(err).Error("error updating entity state in store")
		return err
	}

	event := createKeepaliveEvent(e)
	event.Check.Status = 0
	event.Check.Output = fmt.Sprintf("Keepalive last sent from %s at %s", entity.Name, time.Unix(entity.LastSeen, 0).String())

	return k.bus.Publish(messaging.TopicEventRaw, event)
}
