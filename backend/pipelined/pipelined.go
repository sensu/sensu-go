// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sirupsen/logrus"
)

type WorkflowProcessor struct {
	Filters  []Filter
	Mutators []Mutator
	Handlers []Handler
}

// func NewWorkflowProcessor() WorkflowProcessor {}

//handlers := []Handler{
//	ogPipelineHandler{},
//	tcpStreamHandler{}
//}

type EventProcessor interface {
	Process(context.Context, interface{}) error
}

type Pipeline struct {
	// the pipeline daemon now may be configured with different
	// implementations
	processors []EventProcessor // to be safe, we may want a mutex as well
}

// when we handle an event we iterate through the available processors until
// we find one that can work with the current event.
// func (p *Pipeline) HandleEvent(ctx context.Context, event *corev2.Event) error {
// 	var processor Processor
// 	for _, po := range p.processor {
// 		if !p.processor.CanProcess(event) {
// 			continue
// 		}
// 		processor = po
// 	}
// 	processor.Process(event)
// }

// flatten event handlers (check & metrics), reduced into common interface for pipelines
//workflows := []corev3.PipelineWorkflow{}
//for _, handlers := range event.Check.Handlers {

// check (has both check output and metrics)
//   into pipelined
//     pipelines ->
//       ops-check-pipeline ->
//         -> slack
//       ops-metrics-pipeline ->
//         -> influxdb

var defaultStoreTimeout = time.Minute

// ExtensionExecutorGetterFunc gets an ExtensionExecutor. Used to decouple
// Pipelined from gRPC.
type ExtensionExecutorGetterFunc func(*corev2.Extension) (rpc.ExtensionExecutor, error)

// Pipelined handles incoming Sensu events and puts them through a
// Sensu event pipeline, i.e. filter -> mutator -> handler. The Sensu
// handler configuration determines which Sensu filters and mutator
// are used.
type Pipelined struct {
	assetGetter            asset.Getter
	stopping               chan struct{}
	running                *atomic.Value
	wg                     *sync.WaitGroup
	errChan                chan error
	eventChan              chan interface{}
	subscription           messaging.Subscription
	store                  store.Store
	bus                    messaging.MessageBus
	extensionExecutor      pipeline.ExtensionExecutorGetterFunc
	executor               command.Executor
	workerCount            int
	storeTimeout           time.Duration
	secretsProviderManager *secrets.ProviderManager
	backendEntity          *corev2.Entity
	LicenseGetter          licensing.Getter
}

// Config configures a Pipelined.
type Config struct {
	Store                   store.Store
	Bus                     messaging.MessageBus
	ExtensionExecutorGetter pipeline.ExtensionExecutorGetterFunc
	AssetGetter             asset.Getter
	BufferSize              int
	WorkerCount             int
	StoreTimeout            time.Duration
	SecretsProviderManager  *secrets.ProviderManager
	BackendEntity           *corev2.Entity
	LicenseGetter           licensing.Getter
}

// Option is a functional option used to configure Pipelined.
type Option func(*Pipelined) error

// New creates a new Pipelined with supplied Options applied.
func New(c Config, options ...Option) (*Pipelined, error) {
	if c.BufferSize == 0 {
		logger.Warn("BufferSize not configured")
		c.BufferSize = 1
	}
	if c.WorkerCount == 0 {
		logger.Warn("WorkerCount not configured")
		c.WorkerCount = 1
	}
	if c.StoreTimeout == 0 {
		logger.Warn("StoreTimeout not configured")
		c.StoreTimeout = defaultStoreTimeout
	}

	p := &Pipelined{
		store:                  c.Store,
		bus:                    c.Bus,
		extensionExecutor:      c.ExtensionExecutorGetter,
		stopping:               make(chan struct{}, 1),
		running:                &atomic.Value{},
		wg:                     &sync.WaitGroup{},
		errChan:                make(chan error, 1),
		eventChan:              make(chan interface{}, c.BufferSize),
		workerCount:            c.WorkerCount,
		executor:               command.NewExecutor(),
		assetGetter:            c.AssetGetter,
		storeTimeout:           c.StoreTimeout,
		secretsProviderManager: c.SecretsProviderManager,
		backendEntity:          c.BackendEntity,
		LicenseGetter:          c.LicenseGetter,
	}
	for _, o := range options {
		if err := o(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Receiver returns the event channel for pipelined.
func (p *Pipelined) Receiver() chan<- interface{} {
	return p.eventChan
}

// Start pipelined, subscribing to the "event" message bus topic to
// pass Sensu events to the pipelines for handling (goroutines).
func (p *Pipelined) Start() error {
	sub, err := p.bus.Subscribe(messaging.TopicEvent, "pipelined", p)
	if err != nil {
		return err
	}
	p.subscription = sub

	p.createPipelines(p.workerCount, p.eventChan)

	return nil
}

// Stop pipelined.
func (p *Pipelined) Stop() error {
	p.running.Store(false)
	close(p.stopping)
	p.wg.Wait()
	close(p.errChan)
	err := p.subscription.Cancel()
	close(p.eventChan)

	return err
}

// Err returns a channel to listen for terminal errors on.
func (p *Pipelined) Err() <-chan error {
	return p.errChan
}

// Name returns the daemon name
func (p *Pipelined) Name() string {
	return "pipelined"
}

func (p *Pipelined) pipelineFromEventHandlers(ctx context.Context, event *corev2.Event) (*corev3.Pipeline, error) {
	fields := logrus.Fields{
		"namespace": corev2.ContextNamespace(ctx),
	}

	handlers := []*corev2.Handler{}
	for _, handlerName := range event.Check.Handlers {
		fields["handler"] = handlerName

		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		handler, err := p.store.GetHandlerByName(tctx, handlerName)
		cancel()
		if err != nil {
			if _, ok := err.(*store.ErrInternal); ok {
				// fatal error
				return nil, err
			}
			logger.WithFields(fields).WithError(err).Error("failed to fetch handler")
			continue
		}
		if handler == nil {
			logger.WithFields(fields).WithError(err).Error("fetched handler is nil")
			continue
		}

		handlers = append(handlers, handler)
	}

	return corev3.PipelineFromHandlers(ctx, handlers), nil
}

// createPipelines creates several goroutines, responsible for pulling
// Sensu events from a channel (bound to message bus "event" topic)
// and for handling them.
func (p *Pipelined) createPipelines(count int, channel chan interface{}) {
	for i := 1; i <= count; i++ {
		pipeline := pipeline.New(pipeline.Config{
			Store:                   p.store,
			ExtensionExecutorGetter: p.extensionExecutor,
			AssetGetter:             p.assetGetter,
			StoreTimeout:            p.storeTimeout,
			SecretsProviderManager:  p.secretsProviderManager,
			BackendEntity:           p.backendEntity,
			LicenseGetter:           p.LicenseGetter,
		})
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.stopping:
					return
				case msg := <-channel:
					event, ok := msg.(*corev2.Event)
					if !ok {
						continue
					}

					ctx, cancel := context.WithCancel(context.Background())

					pipelines := []*corev3.Pipeline{}

					// convert event's handlers to their own pipeline and add
					// the pipeline to the list of pipelines.
					handlersPipeline, err := p.pipelineFromEventHandlers(ctx, event)
					if err != nil {
						if _, ok := err.(*store.ErrInternal); ok {
							select {
							case p.errChan <- err:
							case <-p.stopping:
							}
							return
						}
						logger.Error(err)
					}
					pipelines = append(pipelines, handlersPipeline)

					for _, eventPipeline := range pipelines {
						pipeline.
							pipeline.ExecutePipeline()
						if err := pipeline.HandleEvent(ctx, event); err != nil {
						}
						cancel()
						if err != nil {
							if _, ok := err.(*store.ErrInternal); ok {
								select {
								case p.errChan <- err:
								case <-p.stopping:
								}
								return
							}
							logger.Error(err)
						}
					}
				}
			}
		}()
	}
}
