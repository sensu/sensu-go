package metrics

const (
	// StatusLabelName is the name of a label which describes if an operation
	// was successful or not.
	StatusLabelName = "status"

	// StatusLabelSuccess is the value to use for the status label if
	// an operation was successful.
	StatusLabelSuccess = "success"

	// StatusLabelError is the value to use for the status label if
	// an operation resulted in an error.
	StatusLabelError = "error"

	// EventTypeLabelName is the name of a label which describes what type of
	// event the metric is tracking. (e.g. check or metrics)
	EventTypeLabelName = "event_type"

	// EventTypeLabelCheck is the value to use for the event_type label for
	// metrics which represent a check event.
	EventTypeLabelCheck = "check"

	// EventTypeLabelMetrics is the value to use for the event_type label for
	// metrics which represent a metrics event.
	EventTypeLabelMetrics = "metrics"

	// EventTypeLabelUnknown is the value to use for the event_type label for
	// metrics which represent an event where the type is not yet known or
	// an event without both a check & metrics.
	EventTypeLabelUnknown = "unknown"

	// ResourceReferenceLabelName is the name of a label which describes the
	// resource reference a metric is tracking.
	ResourceReferenceLabelName = "resource_ref"
)
