package v2

// A Handler is a handler specification.
#Handler: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the
	// handler
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Type is the handler type, i.e. pipe.
	type?: string @protobuf(2,string)

	// Mutator is the handler event data mutator.
	mutator?: string @protobuf(3,string)

	// Command is the command to be executed for a pipe handler.
	command?: string @protobuf(4,string)

	// Timeout is the handler timeout in seconds.
	timeout?: uint32 @protobuf(5,uint32,#"(gogoproto.jsontag)="timeout""#)

	// Socket contains configuration for a TCP or UDP handler.
	socket?: #HandlerSocket @protobuf(6,HandlerSocket,"(gogoproto.nullable)")

	// Handlers is a list of handlers for a handler set.
	handlers?: [...string] @protobuf(7,string,#"(gogoproto.jsontag)="handlers""#)

	// Filters is a list of filters name to evaluate before executing this
	// handler
	filters?: [...string] @protobuf(8,string,#"(gogoproto.jsontag)="filters""#)

	// EnvVars is a list of environment variables to use with command execution
	envVars?: [...string] @protobuf(9,string,name=env_vars,#"(gogoproto.jsontag)="env_vars""#)

	// RuntimeAssets are a list of assets required to execute a handler.
	runtimeAssets?: [...string] @protobuf(13,string,name=runtime_assets,#"(gogoproto.jsontag)="runtime_assets""#)

	// Secrets is the list of Sensu secrets to set for the handler's
	// execution environment.
	secrets?: [...#Secret] @protobuf(14,Secret,#"(gogoproto.jsontag)="secrets""#)
}

// HandlerSocket contains configuration for a TCP or UDP handler.
#HandlerSocket: {
	// Host is the socket peer address.
	host?: string @protobuf(1,string)

	// Port is the socket peer port.
	port?: uint32 @protobuf(2,uint32,#"(gogoproto.jsontag)="port""#)
}
