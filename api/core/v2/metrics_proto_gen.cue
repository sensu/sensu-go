package v2

// A Metrics is an event metrics payload specification.
#Metrics: {
	// Handlers is a list of handlers for the metric points.
	handlers?: [...string] @protobuf(1,string,#"(gogoproto.jsontag)="handlers""#)

	// Points is a list of metric points (measurements).
	points?: [...#MetricPoint] @protobuf(2,MetricPoint,#"(gogoproto.jsontag)="points""#)
}

// A MetricPoint represents a single measurement.
#MetricPoint: {
	// The metric point name.
	name?: string @protobuf(1,string)

	// The metric point value.
	value?: float64 @protobuf(2,double,#"(gogoproto.jsontag)="value""#)

	// The metric point timestamp, time in nanoseconds since the Epoch.
	timestamp?: int64 @protobuf(3,int64,#"(gogoproto.jsontag)="timestamp""#)

	// Tags is a list of metric tags (dimensions).
	tags?: [...#MetricTag] @protobuf(4,MetricTag,#"(gogoproto.jsontag)="tags""#)
}

// A MetricTag adds a dimension to a metric point.
#MetricTag: {
	// The metric tag name.
	name?: string @protobuf(1,string)

	// The metric tag value.
	value?: string @protobuf(2,string)
}
