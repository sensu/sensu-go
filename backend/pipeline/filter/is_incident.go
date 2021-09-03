package filter

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// IsIncident is a filter adapter which will filter events that are incidents.
type IsIncidentAdapter struct{}

// Name returns the name of the filter adapter.
func (i *IsIncidentAdapter) Name() string {
	return "IsIncidentAdapter"
}

// CanFilter determines whether IsIncidentAdapter can filter the resource being
// referenced.
func (i *IsIncidentAdapter) CanFilter(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" && ref.Name == "is_incident" {
		return true
	}
	return false
}

// Filter will evaluate the event and determine whether or not to filter it.
func (i *IsIncidentAdapter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
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
