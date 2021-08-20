package filter

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// IsIncident is a filter which will filter events that are incidents.
type IsIncident struct{}

// Name returns the name of the pipeline filter.
func (i *IsIncident) Name() string {
	return "IsIncident"
}

// CanFilter determines whether the IsIncident filter can filter the resource
// being referenced.
func (i *IsIncident) CanFilter(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" && ref.Name == "is_incident" {
		return true
	}
	return false
}

// Filter will evaluate the event and determine whether or not to filter it.
func (i *IsIncident) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	// Deny an event if it is neither an incident nor resolution.
	if !event.IsIncident() && !event.IsResolution() {
		logger.WithFields(fields).Debug("denying event that is not an incident/resolution")
		return true, nil
	}

	return false, nil
}
