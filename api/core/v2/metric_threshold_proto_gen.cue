package v2

// A rule to apply to a metric to determine its status
#MetricThresholdRule: {
	// Minimum value for the metric to be considered ok
	min?: string @protobuf(1,string,#"(gogoproto.jsontag)="min""#)

	// Maximum value for the metric to be considered ok
	max?: string @protobuf(2,string,#"(gogoproto.jsontag)="max""#)

	// The status of the metric if the value is below the minimum or above the maximum
	status?: uint32 @protobuf(3,uint32,#"(gogoproto.jsontag)="status""#)

	// The metric status if the measurement is missing
	nullStatus?: uint32 @protobuf(4,uint32,name=null_status,#"(gogoproto.jsontag)="null_status""#)
}

// Represents the measurement tags to match
#MetricThresholdTag: {
	// Name of the metric tag to match
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#)

	// Value of the metric tag to match
	value?: string @protobuf(2,string,#"(gogoproto.jsontag)="value""#)
}

// Represents an instance of a metric filter to evaluate
#MetricThreshold: {
	// Name of the metric to match
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#)

	// Tag values to match with the metric
	tags?: [...#MetricThresholdTag] @protobuf(2,MetricThresholdTag,#"(gogoproto.jsontag)="tags""#)

	// Rules to evaluate when the filter matches a metric
	thresholds?: [...#MetricThresholdRule] @protobuf(3,MetricThresholdRule,#"(gogoproto.jsontag)="thresholds""#)
}
