package v2

// A KeepaliveRecord is a tuple of an entity name and the time at which the
// entity's keepalive will next expire.
#KeepaliveRecord: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name (of the entity), and namespace, labels and
	// annotations of the keepalive record
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")
	time?:     int64       @protobuf(4,int64,#"(gogoproto.jsontag)="time""#)
}
