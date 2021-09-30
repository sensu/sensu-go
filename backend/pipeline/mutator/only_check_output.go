package mutator

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	// OnlyCheckOutputAdapterName is the name of the mutator adapter.
	OnlyCheckOutputAdapterName = "OnlyCheckOutputAdapter"
)

// OnlyCheckOutputAdapter is a mutator adapter which returns only the check
// output from the Sensu event. This mutator is considered to be "built-in"
// (1.x parity), it is most commonly used by tcp/udp handlers (e.g. influxdb).
type OnlyCheckOutputAdapter struct{}

// Name returns the name of the mutator adapter.
func (o *OnlyCheckOutputAdapter) Name() string {
	return OnlyCheckOutputAdapterName
}

// CanMutate determines whether LegacyOnlyCheckOutputAdapter can mutate the
// resource being referenced.
func (o *OnlyCheckOutputAdapter) CanMutate(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" && ref.Name == "only_check_output" {
		return true
	}
	return false
}

// Mutate converts an event's check output to bytes, so long as the event has a
// check, and returns the bytes.
func (o *OnlyCheckOutputAdapter) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	if event.HasCheck() {
		return []byte(event.Check.Output), nil
	}

	return nil, fmt.Errorf("event requires a check for the only_check_output mutator to work, returning")
}
