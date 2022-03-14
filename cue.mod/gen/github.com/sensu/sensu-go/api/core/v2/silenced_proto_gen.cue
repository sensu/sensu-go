package v2

// Silenced is the representation of a silence entry.
#Silenced: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the
	// silenced
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Expire is the number of seconds the entry will live
	expire?: int64 @protobuf(2,int64,"(gogoproto.nullable)",#"(gogoproto.jsontag)="expire""#)

	// ExpireOnResolve defaults to false, clears the entry on resolution when
	// set to true
	expire_on_resolve?: bool @protobuf(3,bool,"(gogoproto.nullable)",#"(gogoproto.jsontag)="expire_on_resolve""#)

	// Creator is the author of the silenced entry
	creator?: string @protobuf(4,string,"(gogoproto.nullable)")

	// Check is the name of the check event to be silenced.
	check?: string @protobuf(5,string)

	// Reason is used to provide context to the entry
	reason?: string @protobuf(6,string,"(gogoproto.nullable)")

	// Subscription is the name of the subscription to which the entry applies.
	subscription?: string @protobuf(7,string,"(gogoproto.nullable)")

	// Begin is a timestamp at which the silenced entry takes effect.
	begin?: int64 @protobuf(10,int64,#"(gogoproto.jsontag)="begin""#)

	// ExpireAt is a timestamp at which the silenced entry will expire.
	expire_at?: int64 @protobuf(11,int64,#"(gogoproto.jsontag)="expire_at""#)
}
