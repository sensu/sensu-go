package mutator

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

// LegacyOnlyCheckOutput is a mutator which returns only the check output from
// the Sensu event. This mutator is considered to be "built-in" (1.x parity), it
// is most commonly used by tcp/udp handlers (e.g. influxdb).
type LegacyOnlyCheckOutput struct{}

// CanMutate determines whether the LegacyOnlyCheckOutput mutator can mutate the
// resource being referenced.
func (l *LegacyOnlyCheckOutput) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" && ref.Name == "only_check_output" {
		return true
	}
	return false
}

// Mutate converts an event's check output to bytes, so long as the event has a
// check, and returns the bytes.
func (l *LegacyOnlyCheckOutput) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)
	logger.WithFields(fields).Info("using built-in only_check_output mutator")

	if event.HasCheck() {
		return []byte(event.Check.Output), nil
	}

	return nil, fmt.Errorf("event requires a check for the only_check_output mutator to work, returning")
}
