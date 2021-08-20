package filter

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// HasMetrics is a filter which will filter events that have metrics.
type HasMetrics struct{}

// Name returns the name of the pipeline filter.
func (i *HasMetrics) Name() string {
	return "HasMetrics"
}

// CanFilter determines whether the HasMetrics filter can filter the resource
// being referenced.
func (i *HasMetrics) CanFilter(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" && ref.Name == "has_metrics" {
		return true
	}
	return false
}

// Filter will evaluate the event and determine whether or not to filter it.
func (i *HasMetrics) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	// Deny an event if it does not have metrics
	if !event.HasMetrics() {
		logger.WithFields(fields).Debug("denying event without metrics")
		return true, nil
	}

	return false, nil
}
