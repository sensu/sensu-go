package v2

// HookConfig is the specification of a hook
#HookConfig: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Metadata contains the name, namespace, labels and annotations of the hook
	metadata?: #ObjectMeta @protobuf(1,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// Command is the command to be executed
	command?: string @protobuf(2,string)

	// Timeout is the timeout, in seconds, at which the hook has to run
	timeout?: uint32 @protobuf(3,uint32,#"(gogoproto.jsontag)="timeout""#)

	// Stdin indicates if hook requests have stdin enabled
	stdin?: bool @protobuf(4,bool,#"(gogoproto.jsontag)="stdin""#)

	// RuntimeAssets are a list of assets required to execute hook.
	runtimeAssets?: [...string] @protobuf(5,string,name=runtime_assets,#"(gogoproto.jsontag)="runtime_assets""#)
}

// A Hook is a hook specification and optionally the results of the hook's
// execution.
#Hook: {
	// Config is the specification of a hook
	config?: #HookConfig @protobuf(1,HookConfig,"(gogoproto.nullable)=false","(gogoproto.embed)",#"(gogoproto.jsontag)="""#)

	// Duration of execution
	duration?: float64 @protobuf(2,double)

	// Executed describes the time in which the hook request was executed
	executed?: int64 @protobuf(3,int64,#"(gogoproto.jsontag)="executed""#)

	// Issued describes the time in which the hook request was issued
	issued?: int64 @protobuf(4,int64,#"(gogoproto.jsontag)="issued""#)

	// Output from the execution of Command
	output?: string @protobuf(5,string)

	// Status is the exit status code produced by the hook
	status?: int32 @protobuf(6,int32,#"(gogoproto.jsontag)="status""#)
}

#HookList: {
	// Hooks is the list of hooks for the check hook
	hooks?: [...string] @protobuf(1,string,#"(gogoproto.jsontag)="hooks""#)

	// Type indicates the type or response code for the check hook
	type?: string @protobuf(2,string)
}
