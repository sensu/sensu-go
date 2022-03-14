package v2

#AdhocRequest: {
	// Subscriptions is the list of entity subscriptions.
	subscriptions?: [...string] @protobuf(2,string,"(gogoproto.nullable)")

	// Creator is the author of the adhoc request.
	creator?: string @protobuf(3,string,"(gogoproto.nullable)")

	// Reason is used to provide context to the request.
	reason?: string @protobuf(4,string,"(gogoproto.nullable)")

	// Metadata contains the name, namespace, labels and annotations of the
	// AdhocCheck
	metadata?: #ObjectMeta @protobuf(5,ObjectMeta,#"(gogoproto.jsontag)="metadata""#,"(gogoproto.embed)","(gogoproto.nullable)=false")
}
