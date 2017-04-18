// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	// PipelineCount specifies how many pipelines (goroutines) are
	// in action.
	PipelineCount int = 10
)

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "store",
	})
}

// Pipelined handles incoming Sensu events and puts them through a
// Sensu event pipeline, i.e. filter -> mutator -> handler. The Sensu
// handler configuration determines which Sensu filters and mutator
// are used.
type Pipelined struct {
	stopping  chan struct{}
	running   *atomic.Value
	wg        *sync.WaitGroup
	errChan   chan error
	eventChan chan []byte

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

	p.eventChan = make(chan []byte, 100)

	if err := p.MessageBus.Subscribe("sensu:event", "pipelined", p.eventChan); err != nil {
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
func (p *Pipelined) createPipelines(count int, channel chan []byte) error {
	for i := 1; i <= count; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.stopping:
					return
				case msg := <-channel:
					event := &types.Event{}
					err := json.Unmarshal(msg, event)

					if err != nil {
						continue
					}

					p.handleEvent(event)
				}
			}
		}()
	}

	return nil
}
