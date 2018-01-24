// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

func evaluateEventFilterStatement(event *types.Event, statement string) bool {
	parameters := map[string]interface{}{"event": event}
	result, err := eval.Evaluate(statement, parameters)
	if err != nil {
		logger.WithError(err).Errorf("statement '%s' is invalid", statement)
		return false
	}

	return result
}

// Returns true if the event should be filtered.
func evaluateEventFilter(event *types.Event, filter *types.EventFilter) bool {
	if filter.When != nil {
		inWindows, err := filter.When.InWindows(time.Now().UTC())
		if err != nil {
			logger.WithField("filter", filter.Name).Error(err)
			return false
		}

		if filter.Action == types.EventFilterActionAllow && !inWindows {
			return true
		}

		if filter.Action == types.EventFilterActionDeny && !inWindows {
			return false
		}
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
	// Do not filter the event if the event has metrics
	if event.HasMetrics() {
		return false
	}

	// Filter if the event has any silenced entries
	if event.IsSilenced() {
		return true
	}

	// Filter the event if it is not an incident and the event has not just
	// transitioned from being an incident to a healthy state
	if !event.IsIncident() && !event.IsResolution() {
		return true
	}

	// Do not filter the event if the handler has no event filters
	if len(handler.Filters) == 0 {
		return false
	}

	// Iterate through all event filters, evaluating each statement against given event. The
	// event is rejected if the product of all statements is true.
	for _, filterName := range handler.Filters {
		// Retrieve the filter from the store with its name
		ctx := types.SetContextFromResource(context.Background(), event.Entity)
		filter, err := p.Store.GetEventFilterByName(ctx, filterName)
		if err != nil {
			logger.WithError(err).Warningf("could not retrieve the filter %s", filterName)
			return false
		}

		filtered := evaluateEventFilter(event, filter)
		if !filtered {
			return false
		}
	}

	return true
}
