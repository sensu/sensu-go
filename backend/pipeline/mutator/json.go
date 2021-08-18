package mutator

import (
	"context"
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// LegacyJSON is a mutator which produces the JSON encoding of the Sensu event.
// This mutator is used when the resource reference refers to a core/v2.Mutator
// with the name, "json".
type LegacyJSON struct{}

// CanMutate determines whether the LegacyJSON mutator can mutate the resource
// being referenced.
func (l *LegacyJSON) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" && ref.Name == "json" {
		return true
	}
	return false
}

// Mutate will mutate the event and return it as JSON encoded bytes.
func (l *LegacyJSON) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)
	logger.WithFields(fields).Info("using built-in json mutator")

	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return eventData, nil
}
