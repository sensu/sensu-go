// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	"github.com/Knetic/govaluate"
)

func evaluateEventFilterStatement(event *types.Event, statement string) bool {
	expr, err := govaluate.NewEvaluableExpression(statement)
	if err != nil {
		logger.WithError(err).Error("failed to parse filter statement: ", statement)
		return false
	}

	result, err := expr.Evaluate(map[string]interface{}{"event": event})
	if err != nil {
		logger.WithError(err).Error("failed to evaluate statement: ", statement)
		return false
	}

	match, ok := result.(bool)
	if !ok {
		logger.WithField("filter", statement).Error("filters must evaluate to boolean values")
	}

	return match
}

// Returns true if the event should be filtered.
func evaluateEventFilter(store store.Store, event *types.Event, filterName string) bool {
	// Retrieve the filter from the store with its name
	ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, event.Entity.Environment)
	filter, err := store.GetEventFilterByName(ctx, filterName)
	if err != nil {
		logger.WithError(err).Warningf("could not retrieve the filter %s", filterName)
		return false
	}

	for _, statement := range filter.Statements {
		match := evaluateEventFilterStatement(event, statement)

		// Allow - One of the statements did not match, filter the event
		if filter.Action == types.EventFilterActionAllow && !match {
			return true
		}

		// Deny - One of the statements did not match, do not filter the event
		if filter.Action == types.EventFilterActionDeny && !match {
			return false
		}
	}

	// Allow - All of the statements matched, do not filter the event
	if filter.Action == types.EventFilterActionAllow {
		return false
	}

	// Deny - All of the statements matched, filter the event
	if filter.Action == types.EventFilterActionDeny {
		return true
	}

	// Something weird happened, let's not filter the event and log a warning message
	logger.WithFields(logrus.Fields{
		"filter":       filter.GetName(),
		"organization": filter.GetOrganization(),
		"environment":  filter.GetEnvironment(),
	}).Warn("pipelined not filtering event due to unhandled case")

	return false
}

// filterEvent filters a Sensu event, determining if it will continue
// through the Sensu pipeline.
func (p *Pipelined) filterEvent(handler *types.Handler, event *types.Event) bool {
	incident := p.isIncident(event)
	metrics := p.hasMetrics(event)

	// Do not filter the event if the event has metrics
	if metrics {
		return false
	}

	// Filter the event if it is not an incident
	if !incident {
		return true
	}

	// Do not filter the event if the handler has no event filters
	if len(handler.Filters) == 0 {
		return false
	}

	// Iterate through all event filters, evaluating each statement against given event. The
	// event is rejected if the product of all statements is true.
	for _, filter := range handler.Filters {
		filtered := evaluateEventFilter(p.Store, event, filter)
		if !filtered {
			return false
		}
	}

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
