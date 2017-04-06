// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"log"

	"github.com/sensu/sensu-go/types"
)

// filterEvent filters a Sensu event, determining if it will continue
// through the Sensu pipeline.
func (p *Pipelined) filterEvent(handler *types.Handler, event *types.Event) bool {
	incident, err := p.isIncident(event)

	if err != nil {
		log.Println("pipelined failed to determine if an event is an incident: ", err.Error())
		return true
	}

	if incident {
		return false
	}

	log.Printf("pipelined filtered an event")

	return true
}

// isIncident determines if an event indicates an incident.
func (p *Pipelined) isIncident(event *types.Event) (bool, error) {
	if event.Check.Status != 0 {
		return true, nil
	}

	return false, nil
}
