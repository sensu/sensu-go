// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

func evaluateEventFilterStatement(event *types.Event, statement string) bool {
	parameters := map[string]interface{}{"event": event}
	result, err := eval.EvaluatePredicate(statement, parameters)
	if err != nil {
		fields := utillogging.EventFields(event, false)
		fields["statement"] = statement
		if _, ok := err.(eval.SyntaxError); ok {
			// Errors during execution are typically due to missing attrs
			logger.WithError(err).WithFields(fields).Error("syntax error")
		} else if _, ok := err.(eval.TypeError); ok {
			logger.WithError(err).WithFields(fields).Error("type error")
		} else {
			logger.WithError(err).WithFields(fields).Debug("missing attribute")
		}
		return false
	}

	return result
}

func evaluateJSFilter(event *types.Event, expr string) bool {
	parameters := map[string]interface{}{"event": event}
	result, err := eval.EvaluateJSExpression(expr, parameters)
	if err != nil {
		logger.WithError(err).Error("error executing JS")
	}
	return result
}

// Returns true if the event should be filtered.
func evaluateEventFilter(event *types.Event, filter *types.EventFilter) bool {
	var evaluator func(*types.Event, string) bool
	if filter.Type == "js" {
		evaluator = evaluateJSFilter
	} else {
		evaluator = evaluateEventFilterStatement
	}

	if filter.When != nil {
		inWindows, err := filter.When.InWindows(time.Now().UTC())
		if err != nil {
			fields := utillogging.EventFields(event, false)
			fields["filter"] = filter.Name
			logger.WithFields(fields).Error(err)
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
		match := evaluator(event, statement)

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
	fields := utillogging.EventFields(event, false)
	fields["filter"] = filter.Name
	logger.WithFields(fields).
		Warn("pipelined not filtering event due to unhandled case")

	return false
}

// filterEvent filters a Sensu event, determining if it will continue
// through the Sensu pipeline.
func (p *Pipelined) filterEvent(handler *types.Handler, event *types.Event) bool {
	// Prepare the logging
	fields := utillogging.EventFields(event, false)
	fields["handler"] = handler.Name

	// Iterate through all event filters, the event is filtered if
	// a filter returns true.
	for _, filterName := range handler.Filters {
		// Do not filter the event if it indicates an incident
		// or incident resolution.
		if filterName == "is_incident" {
			if !event.IsIncident() && !event.IsResolution() {
				return true
			}

			continue
		}

		// Do not filter the event if it has metrics.
		if filterName == "has_metrics" {
			if !event.HasMetrics() {
				return true
			}

			continue
		}

		// Do not filter the event if it is not silenced.
		if filterName == "not_silenced" {
			if event.IsSilenced() {
				return true
			}

			continue
		}

		// Retrieve the filter from the store with its name
		ctx := types.SetContextFromResource(context.Background(), event.Entity)
		filter, err := p.store.GetEventFilterByName(ctx, filterName)
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Warningf("could not retrieve the filter %s", filterName)
			return false
		}

		if filter != nil {
			// Execute the filter, evaluating each of its
			// statements against the event. The event is rejected
			// if the product of all statements is true.
			filtered := evaluateEventFilter(event, filter)
			if filtered {
				return true
			}
			continue
		}

		// If the filter didn't exist, it might be an extension filter
		ext, err := p.store.GetExtension(ctx, filterName)
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Warningf("could not retrieve the filter %s", filterName)
			continue
		}

		executor, err := p.extensionExecutor(ext)
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Errorf("could not execute the filter %s", filterName)
			continue
		}
		defer func() {
			if err := executor.Close(); err != nil {
				logger.WithError(err).Debug("error closing grpc client")
			}
		}()
		filtered, err := executor.FilterEvent(event)
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Errorf("could not execute the filter %s", filterName)
			continue
		}
		if filtered {
			return true
		}
	}

	return false
}
