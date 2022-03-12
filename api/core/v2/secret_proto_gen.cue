package v2

// A Secret is a secret specification.
#Secret: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Name is the name of the secret referenced in an executable command.
	name?: string @protobuf(1,string)

	// Secret is the name of the Sensu secret resource.
	secret?: string @protobuf(2,string)
}
