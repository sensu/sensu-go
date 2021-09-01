package mutator

import (
	"context"
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// JSONAdapter is a mutator adapter which produces the JSON encoding of the
// Sensu event.
type JSONAdapter struct{}

// Name returns the name of the mutator adapter.
func (j *JSONAdapter) Name() string {
	return "JSONAdapter"
}

// CanMutate determines whether JSONAdapter can mutate the resource being
// referenced.
func (j *JSONAdapter) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" && ref.Name == "json" {
		return true
	}
	return false
}

// Mutate will convert the event to JSON and return the data as bytes.
func (j *JSONAdapter) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return eventData, nil
}
