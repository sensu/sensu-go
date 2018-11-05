package logging

import (
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// EventFields populates a map with fields containing relevant information about
// an event for logging
func EventFields(event *types.Event, debug bool) map[string]interface{} {
	// Ensure the entity is present
	if event.Entity == nil {
		return map[string]interface{}{}
	}

	fields := logrus.Fields{
		"entity_id":        event.Entity.ID,
		"entity_namespace": event.Entity.Namespace,
	}

	if event.HasCheck() {
		fields["check_name"] = event.Check.Name
		fields["check_namespace"] = event.Check.Namespace
	}

	if debug {
		fields["timestamp"] = event.Timestamp
		if event.HasMetrics() {
			fields["metrics"] = event.Metrics
		}
		if event.Hooks != nil {
			fields["hooks"] = event.Hooks
		}
		if event.Silenced != nil {
			fields["silenced"] = event.Silenced
		}
	} else {
		if event.HasMetrics() {
			count := len(event.Metrics.Points)
			fields["metric_count"] = count
			if count > 0 {
				fields["first_metric_name"] = event.Metrics.Points[0].Name
				fields["first_metric_value"] = event.Metrics.Points[0].Value
			}
		}
	}

	return fields
}
