package filter

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// HasMetricsAdapter is a filter adapter which will filter events that have
// metrics.
type HasMetricsAdapter struct{}

// Name returns the name of the filter adapter.
func (i *HasMetricsAdapter) Name() string {
	return "HasMetricsAdapter"
}

// CanFilter determines whether HasMetricsAdapter can filter the resource being
// referenced.
func (i *HasMetricsAdapter) CanFilter(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" && ref.Name == "has_metrics" {
		return true
	}
	return false
}

// Filter will evaluate the event and determine whether or not to filter it.
func (i *HasMetricsAdapter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

	// Deny an event if it does not have metrics
	if !event.HasMetrics() {
		logger.WithFields(fields).Debug("denying event without metrics")
		return true, nil
	}

	return false, nil
}
