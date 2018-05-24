package logging

import (
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// EventFields populates a map with fields containing relevant information about
// an event for logging
func EventFields(event *types.Event) map[string]interface{} {
	// Ensure the entity is present
	if event.Entity == nil {
		return map[string]interface{}{}
	}

	fields := logrus.Fields{
		"entity":       event.Entity.ID,
		"environment":  event.Entity.Environment,
		"organization": event.Entity.Organization,
	}

	// The event might not have a check if it's a metric so verify that before
	// adding the check name
	if event.HasCheck() {
		fields["check"] = event.Check.Name
	}

	if event.HasMetrics() {
		count := len(event.Metrics.Points)
		fields["metric_count"] = count
		if count > 0 {
			fields["first_metric_name"] = event.Metrics.Points[0].Name
			fields["first_metric_value"] = event.Metrics.Points[0].Value
		}
	}

	return fields
}
