package v2

// EventFilter is a filter specification.
#EventFilter: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the
	// filter
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Action specifies to allow/deny events to continue through the pipeline
	action?: string @protobuf(2,string)

	// Expressions is an array of boolean expressions that are &&'d together
	// to determine if the event matches this filter.
	expressions?: [...string] @protobuf(3,string,#"(gogoproto.jsontag)="expressions""#)

	// When indicates a TimeWindowWhen that a filter uses to filter by days &
	// times
	when?: #TimeWindowWhen @protobuf(6,TimeWindowWhen)

	// Runtime assets are Sensu assets that contain javascript libraries. They
	// are evaluated within the execution context.
	runtime_assets?: [...string] @protobuf(8,string,#"(gogoproto.jsontag)="runtime_assets""#)
}
