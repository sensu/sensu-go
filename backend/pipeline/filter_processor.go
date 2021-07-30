package pipeline

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
)

type StandardFilterProcessor struct {
	assetGetter       asset.Getter
	extensionExecutor ExtensionExecutorGetterFunc
	store             store.Store
	storeTimeout      time.Duration
}

type StandardMutatorProcessor struct {
	store store.Store
}

func (s *StandardMutatorProcessor) CanMutate(ctx context.Context, ref *corev3.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" {
		return true
	}
	return false
}

func (s *StandardMutatorProcessor) Mutate(ctx context.Context, ref *corev3.ResourceReference, event *corev2.Event) (*corev2.Event, error) {
	return nil, nil
}

type StandardHandlerProcessor struct {
	store store.Store
}

func (s *StandardHandlerProcessor) CanHandle(ctx context.Context, ref *corev3.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Handler" {
		return true
	}
	return false
}

func (s *StandardHandlerProcessor) Handle(ctx context.Context, ref *corev3.ResourceReference, event *corev2.Event) error {
	return nil
}

func (s *StandardFilterProcessor) CanFilter(ctx context.Context, ref *corev3.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" {
		return true
	}
	return false
}

// Filter filters a Sensu event, determining if it will continue through
// the Sensu pipeline. Returns the filter's name if the event was filtered and
// any error encountered
func (s *StandardFilterProcessor) Filter(ctx context.Context, ref *corev3.ResourceReference, event *corev2.Event) (bool, error) {
	switch ref.Name {
	case "is_incident":
		// Deny an event if it is neither an incident nor resolution.
		if !event.IsIncident() && !event.IsResolution() {
			//logger.WithFields(fields).Debug("denying event that is not an incident/resolution")
			return true, nil
		}
	case "has_metrics":
		// Deny an event if it does not have metrics
		if !event.HasMetrics() {
			//logger.WithFields(fields).Debug("denying event without metrics")
			return true, nil
		}
	case "not_silenced":
		// Deny event that is silenced.
		if event.IsSilenced() {
			//logger.WithFields(fields).Debug("denying event that is silenced")
			return true, nil
		}
	default:
		// Retrieve the filter from the store with its name
		ctx := corev2.SetContextFromResource(context.Background(), event.Entity)
		tctx, cancel := context.WithTimeout(ctx, s.storeTimeout)
		filter, err := s.store.GetEventFilterByName(tctx, ref.Name)
		cancel()
		if err != nil {
			//logger.WithFields(fields).WithError(err).
			//	Warning("could not retrieve filter")
			return false, err
		}

		if filter != nil {
			// Execute the filter, evaluating each of its
			// expressions against the event. The event is rejected
			// if the product of all expressions is true.
			ctx := corev2.SetContextFromResource(context.Background(), filter)
			matchedAssets := asset.GetAssets(ctx, s.store, filter.RuntimeAssets)
			assets, err := asset.GetAll(ctx, s.assetGetter, matchedAssets)
			if err != nil {
				//logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for filter")
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					return false, err
				}
			}
			filtered := evaluateEventFilter(ctx, event, filter, assets)
			if filtered {
				//logger.WithFields(fields).Debug("denying event with custom filter")
				return true, nil
			}
			return false, nil
		}

		// If the filter didn't exist, it might be an extension filter
		ext, err := s.store.GetExtension(ctx, ref.Name)
		if err != nil {
			//logger.WithFields(fields).WithError(err).
			//	Warning("could not retrieve filter")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return false, err
			}
			return false, nil
		}

		executor, err := s.extensionExecutor(ext)
		if err != nil {
			//logger.WithFields(fields).WithError(err).
			//	Error("could not execute filter")
			return false, nil
		}
		defer func() {
			if err := executor.Close(); err != nil {
				logger.WithError(err).Debug("error closing grpc client")
			}
		}()
		filtered, err := executor.FilterEvent(event)
		if err != nil {
			//logger.WithFields(fields).WithError(err).
			//	Error("could not execute filter")
			return false, nil
		}
		if filtered {
			//logger.WithFields(fields).Debug("denying event with custom filter extension")
			return true, nil
		}
	}

	//logger.WithFields(fields).Debug("allowing event")
	return false, nil
}
