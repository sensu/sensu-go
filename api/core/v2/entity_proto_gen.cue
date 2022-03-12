package v2

// Entity is the Entity supplying the event. The default Entity for any
// Event is the running Agent process--if the Event is sent by an Agent.
#Entity: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)
	entityClass?: string  @protobuf(1,string,name=entity_class,#"(gogoproto.jsontag)="entity_class""#)
	system?:      #System @protobuf(3,System,"(gogoproto.nullable)=false")
	subscriptions?: [...string] @protobuf(4,string,#"(gogoproto.jsontag)="subscriptions""#)
	lastSeen?:       int64           @protobuf(5,int64,name=last_seen,#"(gogoproto.jsontag)="last_seen""#)
	deregister?:     bool            @protobuf(6,bool,#"(gogoproto.jsontag)="deregister""#)
	deregistration?: #Deregistration @protobuf(7,Deregistration,"(gogoproto.nullable)=false")
	user?:           string          @protobuf(11,string)

	// ExtendedAttributes store serialized arbitrary JSON-encoded data
	extendedAttributes?: bytes @protobuf(12,bytes,name=extended_attributes,#"(gogoproto.jsontag)="-""#)

	// Redact contains the fields to redact on the agent
	redact?: [...string] @protobuf(13,string)

	// Metadata contains the name, namespace, labels and annotations of the
	// entity
	metadata?: #ObjectMeta @protobuf(14,ObjectMeta,#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.embed)","(gogoproto.nullable)=false")

	// SensuAgentVersion is the sensu-agent version.
	sensuAgentVersion?: string @protobuf(15,string,name=sensu_agent_version,#"(gogoproto.jsontag)="sensu_agent_version""#)

	// KeepaliveHandlers contains a list of handlers to use for the entity's
	// keepalive events
	keepaliveHandlers?: [...string] @protobuf(16,string,name=keepalive_handlers)
}

// System contains information about the system that the Agent process
// is running on, used for additional Entity context.
#System: {
	hostname?:        string   @protobuf(1,string)
	os?:              string   @protobuf(2,string,#"(gogoproto.customname)="OS""#)
	platform?:        string   @protobuf(3,string)
	platformFamily?:  string   @protobuf(4,string,name=platform_family)
	platformVersion?: string   @protobuf(5,string,name=platform_version)
	network?:         #Network @protobuf(6,Network,"(gogoproto.nullable)=false")
	arch?:            string   @protobuf(7,string)
	armVersion?:      int32    @protobuf(8,int32,name=arm_version,#"(gogoproto.customname)="ARMVersion""#)

	// LibCType indicates the type of libc the agent has access to (glibc, musl,
	// etc)
	LibCType?: string @protobuf(9,string,#"(gogoproto.jsontag)="libc_type""#)

	// VMSystem indicates the VM system of the agent (kvm, vbox, etc)
	VMSystem?: string @protobuf(10,string,#"(gogoproto.jsontag)="vm_system""#)

	// VMRole indicates the VM role of the agent (host/guest)
	VMRole?: string @protobuf(11,string,#"(gogoproto.jsontag)="vm_role""#)

	// CloudProvider indicates the public cloud the agent is running on.
	CloudProvider?: string @protobuf(12,string,#"(gogoproto.jsontag)="cloud_provider""#)
	floatType?:     string @protobuf(13,string,name=float_type)

	// Processes contains information about the local processes on the agent.
	Processes?: [...#Process] @protobuf(14,Process,#"(gogoproto.jsontag)="processes""#)
}

// Process contains information about a local process.
#Process: {
	name?: string @protobuf(1,string,#"(gogoproto.jsontag)="name""#)
}

// Network contains information about the system network interfaces
// that the Agent process is running on, used for additional Entity
// context.
#Network: {
	interfaces?: [...#NetworkInterface] @protobuf(1,NetworkInterface,#"(gogoproto.jsontag)="interfaces""#,"(gogoproto.nullable)=false")
}

// NetworkInterface contains information about a system network
// interface.
#NetworkInterface: {
	name?: string @protobuf(1,string)
	mac?:  string @protobuf(2,string,#"(gogoproto.customname)="MAC""#)
	addresses?: [...string] @protobuf(3,string,#"(gogoproto.jsontag)="addresses""#)
}

// Deregistration contains configuration for Sensu entity de-registration.
#Deregistration: {
	handler?: string @protobuf(1,string)
}
