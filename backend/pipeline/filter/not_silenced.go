package filter

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// NotSilencedAdapter is a filter adapter which will filter events that are not
// silenced.
type NotSilencedAdapter struct{}

// Name returns the name of the filter adapter.
func (n *NotSilencedAdapter) Name() string {
	return "NotSilencedAdapter"
}

// CanFilter determines whether NotSilencedAdapter can filter the resource being
// referenced.
func (n *NotSilencedAdapter) CanFilter(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" && ref.Name == "not_silenced" {
		return true
	}
	return false
}

// Filter will evaluate the event and determine whether or not to filter it.
func (n *NotSilencedAdapter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	// Deny an event if it is silenced
	if event.IsSilenced() {
		logger.WithFields(fields).Debug("denying event that is silenced")
		return true, nil
	}

	return false, nil
}
