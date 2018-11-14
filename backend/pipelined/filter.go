// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"context"
	"time"

	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

func evaluateJSFilter(event interface{}, expr string, assets asset.RuntimeAssetSet) bool {
	parameters := map[string]interface{}{"event": event}
	result, err := js.Evaluate(expr, parameters, assets)
	if err != nil {
		logger.WithError(err).Error("error executing JS")
	}
	return result
}

// Returns true if the event should be filtered/denied.
func evaluateEventFilter(event *types.Event, filter *types.EventFilter, assets asset.RuntimeAssetSet) bool {
	fields := utillogging.EventFields(event, false)
	fields["filter"] = filter.Name
	fields["assets"] = filter.RuntimeAssets

	if filter.When != nil {
		inWindows, err := filter.When.InWindows(time.Now().UTC())
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Error("denying event - unable to determine if time is in specified window")
			return false
		}

		if filter.Action == types.EventFilterActionAllow && !inWindows {
			logger.WithFields(fields).Debug("denying event outside of filtering window")
			return true
		}

		if filter.Action == types.EventFilterActionDeny && !inWindows {
			logger.WithFields(fields).Debug("allowing event outside of filtering window")
			return false
		}
	}

	synth := dynamic.Synthesize(event)

	for _, expression := range filter.Expressions {
		match := evaluateJSFilter(synth, expression, assets)

		// Allow - One of the expressions did not match, filter the event
		if filter.Action == types.EventFilterActionAllow && !match {
			logger.WithFields(fields).Debug("denying event that does not match filter")
			return true
		}

		// Deny - One of the expressions did not match, do not filter the event
		if filter.Action == types.EventFilterActionDeny && !match {
			logger.WithFields(fields).Debug("allowing event that does not match filter")
			return false
		}
	}

	// Allow - All of the expressions matched, do not filter the event
	if filter.Action == types.EventFilterActionAllow {
		logger.WithFields(fields).Debug("allowing event that matches filter")
		return false
	}

	// Deny - All of the expressions matched, filter the event
	if filter.Action == types.EventFilterActionDeny {
		logger.WithFields(fields).Debug("denying event that matches filter")
		return true
	}

	// Something weird happened, let's not filter the event and log a warning message
	logger.WithFields(fields).
		Warn("not filtering event due to unhandled case")

	return false
}

// filterEvent filters a Sensu event, determining if it will continue
// through the Sensu pipeline. Returns true if the event should be filtered/denied.
func (p *Pipelined) filterEvent(handler *types.Handler, event *types.Event) bool {
	// Prepare the logging
	fields := utillogging.EventFields(event, false)
	fields["handler"] = handler.Name

	// Iterate through all event filters, the event is filtered if
	// a filter returns true.
	for _, filterName := range handler.Filters {
		fields["filter"] = filterName

		switch filterName {
		case "is_incident":
			// Deny an event if it is neither an incident nor resolution.
			if !event.IsIncident() && !event.IsResolution() {
				logger.WithFields(fields).Debug("denying event that is not an incident/resolution")
				return true
			}
		case "has_metrics":
			// Deny an event if it does not have metrics
			if !event.HasMetrics() {
				logger.WithFields(fields).Debug("denying event without metrics")
				return true
			}
		case "not_silenced":
			// Deny event that is silenced.
			if event.IsSilenced() {
				logger.WithFields(fields).Debug("denying event that is silenced")
				return true
			}
		default:
			// Retrieve the filter from the store with its name
			ctx := types.SetContextFromResource(context.Background(), event.Entity)
			filter, err := p.store.GetEventFilterByName(ctx, filterName)
			if err != nil {
				logger.WithFields(fields).WithError(err).
					Warning("could not retrieve filter")
				return false
			}

			if filter != nil {
				// Execute the filter, evaluating each of its
				// expressions against the event. The event is rejected
				// if the product of all expressions is true.
				ctx := types.SetContextFromResource(context.Background(), filter)
				matchedAssets := asset.GetAssets(ctx, p.store, filter.RuntimeAssets)
				assets, err := asset.GetAll(p.assetGetter, matchedAssets)
				if err != nil {
					logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for filter")
				}
				filtered := evaluateEventFilter(event, filter, assets)
				if filtered {
					logger.WithFields(fields).Debug("denying event with custom filter")
					return true
				}
				continue
			}

			// If the filter didn't exist, it might be an extension filter
			ext, err := p.store.GetExtension(ctx, filterName)
			if err != nil {
				logger.WithFields(fields).WithError(err).
					Warning("could not retrieve filter")
				continue
			}

			executor, err := p.extensionExecutor(ext)
			if err != nil {
				logger.WithFields(fields).WithError(err).
					Error("could not execute filter")
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
					Error("could not execute filter")
				continue
			}
			if filtered {
				logger.WithFields(fields).Debug("denying event with custom filter extension")
				return true
			}
		}
	}

	logger.WithFields(fields).Debug("allowing event")
	return false
}
