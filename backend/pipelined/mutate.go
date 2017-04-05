// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"errors"

	"github.com/sensu/sensu-go/types"
)

func (p *Pipelined) mutateEvent(handler *types.Handler, event *types.Event) ([]byte, error) {
	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)

		return eventData, err
	}

	// We should guard against creating a handler with an
	// unknown (missing) mutator.
	return nil, errors.New("unknown mutator")
}

func (p *Pipelined) jsonMutator(event *types.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)

	return eventData, err
}
