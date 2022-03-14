package v2

// A Mutator is a mutator specification.
#Mutator: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the
	// mutator
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Command is the command to be executed.
	command?: string @protobuf(2,string)

	// Timeout is the command execution timeout in seconds.
	timeout?: uint32 @protobuf(3,uint32,#"(gogoproto.jsontag)="timeout""#)

	// Env is a list of environment variables to use with command execution
	env_vars?: [...string] @protobuf(4,string,#"(gogoproto.jsontag)="env_vars""#)

	// RuntimeAssets are a list of assets required to execute a mutator.
	runtime_assets?: [...string] @protobuf(8,string,#"(gogoproto.jsontag)="runtime_assets""#)

	// Secrets is the list of Sensu secrets to set for the mutators's
	// execution environment.
	secrets?: [...#Secret] @protobuf(9,Secret,#"(gogoproto.jsontag)="secrets""#)

	// Type specifies the type of the mutator. If blank or set to "pipe", the
	// mutator will execute a command with the default shell of the sensu user.
	// If set to "javascript", the eval field will be used, interpreted as ECMAScript 5
	// and run on the Otto VM. The runtime assets will be assumed to be javascript
	// assets, and the environment variables will be made available to the global
	// environment of the mutator.
	type?: string @protobuf(10,string,#"(gogoproto.jsontag)="type,omitempty""#)

	// When the type of the mutator is "javascript", the eval field will be expected
	// to hold a valid ECMAScript 5 expression.
	eval?: string @protobuf(11,string,#"(gogoproto.jsontag)="eval,omitempty""#)
}
