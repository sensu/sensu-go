package v2

// An Event is the encapsulating type sent across the Sensu websocket transport.
#Event: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Timestamp is the time in seconds since the Epoch.
	timestamp?: int64 @protobuf(1,int64)

	// Entity describes the entity in which the event occurred.
	entity?: #Entity @protobuf(2,Entity,"(gogoproto.nullable)")

	// Check describes the result of a check; if event is associated to check
	// execution.
	check?: #Check @protobuf(3,Check,"(gogoproto.nullable)")

	// Metrics are zero or more Sensu metrics
	metrics?: #Metrics @protobuf(4,Metrics,"(gogoproto.nullable)")

	// Metadata contains name, namespace, labels and annotations
	metadata?: #ObjectMeta @protobuf(5,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata""#,"(gogoproto.nullable)=false")

	// ID is the unique identifier of the event.
	ID?: bytes @protobuf(6,bytes,#"(gogoproto.jsontag)="id""#)

	// Sequence is the event sequence number. The agent increments the sequence
	// number by one for every successive event. When the agent restarts or
	// reconnects to another backend, the sequence number is reset to 1.
	Sequence?: int64 @protobuf(7,int64,#"(gogoproto.jsontag)="sequence""#)

	// Pipelines are the pipelines that should be used to process an event.
	// APIVersion should default to "core/v2" and Type should default to
	// "Pipeline".
	pipelines?: [...#ResourceReference] @protobuf(8,ResourceReference,#"(gogoproto.jsontag)="pipelines""#)
}
