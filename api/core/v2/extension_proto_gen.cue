package v2

// Extension is a registered sensu extension.
#Extension: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the
	// extension
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// URL is the URL of the gRPC service that implements the extension.
	url?: string @protobuf(2,string,#"(gogoproto.customname)="URL""#)
}
