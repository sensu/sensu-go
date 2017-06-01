// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"errors"
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

	Store      store.Store
	MessageBus messaging.MessageBus
}

// Start pipelined, subscribing to the "event" message bus topic to
// pass Sensu events to the pipelines for handling (goroutines).
func (p *Pipelined) Start() error {
	if p.Store == nil {
		return errors.New("no store found")
	}

	if p.MessageBus == nil {
		return errors.New("no message bus found")
	}

	p.stopping = make(chan struct{}, 1)
	p.running = &atomic.Value{}
	p.wg = &sync.WaitGroup{}

	p.errChan = make(chan error, 1)

	p.eventChan = make(chan interface{}, 100)

	if err := p.MessageBus.Subscribe(messaging.TopicEvent, "pipelined", p.eventChan); err != nil {
		return err
	}

	if err := p.createPipelines(PipelineCount, p.eventChan); err != nil {
		return err
	}

	return nil
}

// Stop pipelined.
func (p *Pipelined) Stop() error {
	p.running.Store(false)
	close(p.stopping)
	p.wg.Wait()
	close(p.errChan)
	// eventChan is closed by MessageBus Stop()

	return nil
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
func (p *Pipelined) createPipelines(count int, channel chan interface{}) error {
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

					p.handleEvent(event)
				}
			}
		}()
	}

	return nil
}
