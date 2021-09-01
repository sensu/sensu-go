// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/pipeline"
	"github.com/sensu/sensu-go/backend/store"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var defaultStoreTimeout = time.Minute

// Pipelined handles incoming Sensu events and puts them through a
// Sensu event pipeline, i.e. filter -> mutator -> handler. The Sensu
// handler configuration determines which Sensu filters and mutator
// are used.
type Pipelined struct {
	stopping     chan struct{}
	running      *atomic.Value
	wg           *sync.WaitGroup
	errChan      chan error
	eventChan    chan interface{}
	subscription messaging.Subscription
	bus          messaging.MessageBus
	workerCount  int
	store        store.Store
	storeTimeout time.Duration
	adapters     []pipeline.Adapter
}

// Config configures a Pipelined.
type Config struct {
	Bus          messaging.MessageBus
	BufferSize   int
	Store        store.Store
	StoreTimeout time.Duration
	WorkerCount  int
}

// Option is a functional option used to configure Pipelined.
type Option func(*Pipelined) error

// PipelineGetter defines an interface for any structures which can return a
// slice of Pipeline resource references.
type PipelineGetter interface {
	GetPipelines() []*corev2.ResourceReference
}

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
		bus:          c.Bus,
		stopping:     make(chan struct{}, 1),
		running:      &atomic.Value{},
		wg:           &sync.WaitGroup{},
		errChan:      make(chan error, 1),
		eventChan:    make(chan interface{}, c.BufferSize),
		workerCount:  c.WorkerCount,
		store:        c.Store,
		storeTimeout: c.StoreTimeout,
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

	p.createWorkers(p.workerCount, p.eventChan)

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

func (p *Pipelined) AddAdapter(adapter pipeline.Adapter) {
	p.adapters = append(p.adapters, adapter)
}

// createWorkers creates several goroutines, responsible for pulling
// Sensu events from a channel (bound to message bus "event" topic)
// and passing them to their referenced pipelines.
func (p *Pipelined) createWorkers(count int, channel chan interface{}) {
	for i := 1; i <= count; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.stopping:
					return
				case msg := <-channel:
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					if err := p.handleMessage(ctx, msg); err != nil {
						if _, ok := err.(*store.ErrInternal); ok {
							select {
							case p.errChan <- err:
							case <-p.stopping:
							}
							return
						}
					}
				}
			}
		}()
	}
}

func (p *Pipelined) handleMessage(ctx context.Context, msg interface{}) error {
	getter, ok := msg.(PipelineGetter)
	if !ok {
		panic("message received was not a PipelineGetter")
	}

	pipelineRefs := getter.GetPipelines()

	// Add a legacy pipeline "reference" if msg is a
	// corev2.Event & has handlers.
	if event, ok := msg.(*corev2.Event); ok {
		// Prepare log entry
		fields := utillogging.EventFields(event, false)
		if event.HasHandlers() {
			legacyPipelineRef := &corev2.ResourceReference{
				APIVersion: "core/v2",
				Type:       "Pipeline",
				Name:       pipeline.LegacyPipelineName,
			}
			pipelineRefs = append(pipelineRefs, legacyPipelineRef)
		} else {
			logger.WithFields(fields).Debug("event has no handlers defined, skipping addition of legacy pipeline reference")
		}
	}

	if len(pipelineRefs) == 0 {
		//logger.WithFields(fields).Info("no pipelines defined")
		return nil
	}

	// loop through list of pipeline references and find
	// adapters that can run each of them.
	for _, ref := range pipelineRefs {
		adapterFound := false

		fields := logrus.Fields{
			"pipeline_reference": ref.ResourceID(),
		}

		for _, adapter := range p.adapters {
			if adapter.CanRun(ctx, ref) {
				if err := adapter.Run(ctx, ref, msg); err != nil {
					if _, ok := err.(*store.ErrInternal); ok {
						return err
					}
					logger.WithFields(fields).Error(err)
				}
				adapterFound = true
			}
		}
		if !adapterFound {
			return fmt.Errorf("no pipeline adapters were found that support the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
		}
	}

	return nil
}
