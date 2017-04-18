// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"github.com/sensu/sensu-go/types"
)

// filterEvent filters a Sensu event, determining if it will continue
// through the Sensu pipeline.
func (p *Pipelined) filterEvent(handler *types.Handler, event *types.Event) bool {
	incident := p.isIncident(event)

	metrics := p.hasMetrics(event)

	if incident || metrics {
		return false
	}

	logger.Debug("pipelined filtered an event")

	return true
}

// isIncident determines if an event indicates an incident.
func (p *Pipelined) isIncident(event *types.Event) bool {
	if event.Check.Status != 0 {
		return true
	}

	return false
}

// hasMetrics determines if an event has metric data.
func (p *Pipelined) hasMetrics(event *types.Event) bool {
	if event.Metrics != nil {
		return true
	}

	return false
}
