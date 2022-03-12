package v2

// An APIKey is an api key specification.
#APIKey: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace (N/A), labels and annotations of
	// the APIKey.
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Username is the username associated with the API key.
	username?: string @protobuf(2,string)

	// CreatedAt is a timestamp which the API key was created.
	createdAt?: int64 @protobuf(3,int64,name=created_at)
}
