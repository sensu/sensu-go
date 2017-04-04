// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/types"
)

const (
	PipelineCount int = 10
)

// Pipelined handles incoming Sensu events and puts them through the
// Sensu event pipeline, i.e. filter -> mutator -> handler.
type Pipelined struct {
	stopping  chan struct{}
	running   *atomic.Value
	wg        *sync.WaitGroup
	errChan   chan error
	eventChan chan []byte

	MessageBus messaging.MessageBus
}

// Start pipelined.
func (p *Pipelined) Start() error {
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
	close(p.eventChan)

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

func (p *Pipelined) handleEvent(event *types.Event) error {
	eventData, err := p.executeEventMutator(event)

	if err != nil {
		return nil
	}

	handlerExec, err := p.executeEventHandler(eventData)

	if err != nil {
		return nil
	}

	log.Printf("executed event handler: status: %x output: %s", handlerExec.Status, handlerExec.Output)

	return err
}

func (p *Pipelined) executeEventHandler(eventData []byte) (*command.Execution, error) {
	handlerExec := &command.Execution{}

	handlerExec.Command = "cat"
	handlerExec.Input = string(eventData[:])

	result, err := command.ExecuteCommand(context.Background(), handlerExec)

	return result, err
}

func (p *Pipelined) executeEventMutator(event *types.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)

	return eventData, err
}
