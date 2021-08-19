package mutator

import (
	"context"
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// JSON is a mutator which produces the JSON encoding of the Sensu event.
type JSON struct{}

// Name returns the name of the pipeline mutator.
func (j *JSON) Name() string {
	return "JSON"
}

// CanMutate determines whether the JSON mutator can mutate the resource being
// referenced.
func (j *JSON) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" && ref.Name == "json" {
		return true
	}
	return false
}

// Mutate will convert the event to JSON and return the data as bytes.
func (j *JSON) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return eventData, nil
}
