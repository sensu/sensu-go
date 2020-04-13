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
		"entity_name":      event.Entity.Name,
		"entity_namespace": event.Entity.Namespace,
		"uuid":             event.GetUUID().String(),
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
		if event.HasCheck() {
			fields["hooks"] = event.Check.Hooks
			fields["silenced"] = event.Check.Silenced
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

// IncrementLogLevel increments the log level, or wraps around to "error" level.
func IncrementLogLevel(level logrus.Level) logrus.Level {
	level++
	if level > logrus.TraceLevel {
		// wrap around to error level
		level = logrus.ErrorLevel
	}
	return level
}
