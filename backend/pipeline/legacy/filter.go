package legacy

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types/dynamic"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

type Filter struct {
	AssetGetter  asset.Getter
	Store        store.Store
	StoreTimeout time.Duration
}

func (f *Filter) CanFilter(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" {
		return true
	}
	return false
}

// Filter filters a Sensu event, determining if it will continue through the
// Sensu pipeline. It returns whether or not the event was filtered and if any
// error was encountered.
func (f *Filter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	switch ref.Name {
	case "is_incident":
		// Deny an event if it is neither an incident nor resolution.
		if !event.IsIncident() && !event.IsResolution() {
			logger.WithFields(fields).Debug("denying event that is not an incident/resolution")
			return true, nil
		}
	case "has_metrics":
		// Deny an event if it does not have metrics
		if !event.HasMetrics() {
			logger.WithFields(fields).Debug("denying event without metrics")
			return true, nil
		}
	case "not_silenced":
		// Deny event that is silenced.
		if event.IsSilenced() {
			logger.WithFields(fields).Debug("denying event that is silenced")
			return true, nil
		}
	default:
		// Retrieve the filter from the store with its name
		ctx := corev2.SetContextFromResource(context.Background(), event.Entity)
		tctx, cancel := context.WithTimeout(ctx, f.StoreTimeout)
		filter, err := f.Store.GetEventFilterByName(tctx, ref.Name)
		cancel()
		if err != nil {
			logger.WithFields(fields).WithError(err).Warning("could not retrieve filter")
			return false, err
		}

		if filter != nil {
			// Execute the filter, evaluating each of its
			// expressions against the event. The event is rejected
			// if the product of all expressions is true.
			ctx := corev2.SetContextFromResource(context.Background(), filter)
			matchedAssets := asset.GetAssets(ctx, f.Store, filter.RuntimeAssets)
			assets, err := asset.GetAll(ctx, f.AssetGetter, matchedAssets)
			if err != nil {
				logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for filter")
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					return false, err
				}
			}
			filtered := evaluateEventFilter(ctx, event, filter, assets)
			if filtered {
				logger.WithFields(fields).Debug("denying event with custom filter")
				return true, nil
			}
			return false, nil
		}

		// TODO: handle if the filter didn't exist
		logger.WithFields(fields).WithError(err).Warning("could not retrieve filter")
	}

	logger.WithFields(fields).Debug("allowing event")
	return false, nil
}

// Returns true if the event should be filtered/denied.
func evaluateEventFilter(ctx context.Context, event *corev2.Event, filter *corev2.EventFilter, assets asset.RuntimeAssetSet) bool {
	// Redact the entity to avoid leaking sensitive information
	event.Entity = event.Entity.GetRedactedEntity()

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

		if filter.Action == corev2.EventFilterActionAllow && !inWindows {
			logger.WithFields(fields).Debug("denying event outside of filtering window")
			return true
		}

		if filter.Action == corev2.EventFilterActionDeny && !inWindows {
			logger.WithFields(fields).Debug("allowing event outside of filtering window")
			return false
		}
	}

	// Guard against nil metadata labels and annotations to improve the user
	// experience of querying these them.
	if event.ObjectMeta.Annotations == nil {
		event.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.ObjectMeta.Labels == nil {
		event.ObjectMeta.Labels = make(map[string]string)
	}
	if event.Check.ObjectMeta.Annotations == nil {
		event.Check.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.Check.ObjectMeta.Labels == nil {
		event.Check.ObjectMeta.Labels = make(map[string]string)
	}
	if event.Entity.ObjectMeta.Annotations == nil {
		event.Entity.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.Entity.ObjectMeta.Labels == nil {
		event.Entity.ObjectMeta.Labels = make(map[string]string)
	}

	synth := dynamic.Synthesize(event)
	env := FilterExecutionEnvironment{
		Event:  synth,
		Assets: assets,
		Funcs:  PipelineFilterFuncs,
	}

	for _, expression := range filter.Expressions {
		match, err := env.Eval(ctx, expression)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("error evaluating javascript event filter")
			continue
		}

		// Allow - One of the expressions did not match, filter the event
		if filter.Action == corev2.EventFilterActionAllow && !match {
			logger.WithFields(fields).Debug("denying event that does not match filter")
			return true
		}

		// Deny - One of the expressions did not match, do not filter the event
		if filter.Action == corev2.EventFilterActionDeny && !match {
			logger.WithFields(fields).Debug("allowing event that does not match filter")
			return false
		}
	}

	// Allow - All of the expressions matched, do not filter the event
	if filter.Action == corev2.EventFilterActionAllow {
		logger.WithFields(fields).Debug("allowing event that matches filter")
		return false
	}

	// Deny - All of the expressions matched, filter the event
	if filter.Action == corev2.EventFilterActionDeny {
		logger.WithFields(fields).Debug("denying event that matches filter")
		return true
	}

	// Something weird happened, let's not filter the event and log a warning message
	logger.WithFields(fields).
		Warn("not filtering event due to unhandled case")

	return false
}
