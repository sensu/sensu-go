// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	// PipelineCount specifies how many pipelines (goroutines) are
	// in action.
	PipelineCount int = 10
)

// Pipelined handles incoming Sensu events and puts them through a
// Sensu event pipeline, i.e. filter -> mutator -> handler. The Sensu
// handler configuration determines which Sensu filters and mutator
// are used.
type Pipelined struct {
	stopping  chan struct{}
	running   *atomic.Value
	wg        *sync.WaitGroup
	errChan   chan error
	eventChan chan interface{}
	store     store.Store
	bus       messaging.MessageBus
}

// Config configures a Pipelined.
type Config struct {
	Store store.Store
	Bus   messaging.MessageBus
}

// Option is a functional option used to configure Pipelined.
type Option func(*Pipelined) error

// New creates a new Pipelined with supplied Options applied.
func New(c Config, options ...Option) (*Pipelined, error) {
	p := &Pipelined{
		store:     c.Store,
		bus:       c.Bus,
		stopping:  make(chan struct{}, 1),
		running:   &atomic.Value{},
		wg:        &sync.WaitGroup{},
		errChan:   make(chan error, 1),
		eventChan: make(chan interface{}, 100),
	}
	for _, o := range options {
		if err := o(p); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// Start pipelined, subscribing to the "event" message bus topic to
// pass Sensu events to the pipelines for handling (goroutines).
func (p *Pipelined) Start() error {
	if err := p.bus.Subscribe(messaging.TopicEvent, "pipelined", p.eventChan); err != nil {
		return err
	}

	p.createPipelines(PipelineCount, p.eventChan)

	return nil
}

// Stop pipelined.
func (p *Pipelined) Stop() error {
	p.running.Store(false)
	close(p.stopping)
	p.wg.Wait()
	close(p.errChan)
	err := p.bus.Unsubscribe(messaging.TopicEvent, "pipelined")
	close(p.eventChan)

	return err
}

// Status returns an error if pipelined is unhealthy.
func (p *Pipelined) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (p *Pipelined) Err() <-chan error {
	return p.errChan
}

// createPipelines creates several goroutines, responsible for pulling
// Sensu events from a channel (bound to message bus "event" topic)
// and for handling them.
func (p *Pipelined) createPipelines(count int, channel chan interface{}) {
	for i := 1; i <= count; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.stopping:
					return
				case msg := <-channel:
					event, ok := msg.(*types.Event)
					if !ok {
						continue
					}

					if err := p.handleEvent(event); err != nil {
						logger.Error(err)
					}
				}
			}
		}()
	}
}
