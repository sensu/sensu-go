package v2

// Rule holds information that describes an action that can be taken
#Rule: {
	// Verbs is a list of verbs that apply to all of the listed resources for
	// this rule. These include "get", "list", "watch", "create", "update",
	// "delete".
	// TODO: add support for "patch" (this is expensive and should be delayed
	// until a further release). TODO: add support for "watch" (via websockets)
	verbs?: [...string] @protobuf(1,string,#"(gogoproto.jsontag)="verbs""#)

	// Resources is a list of resources that this rule applies to. "*"
	// represents all resources.
	resources?: [...string] @protobuf(2,string,#"(gogoproto.jsontag)="resources""#)

	// ResourceNames is an optional list of resource names that the rule applies
	// to.
	resourceNames?: [...string] @protobuf(3,string,name=resource_names,#"(gogoproto.jsontag)="resource_names""#)
}

// ClusterRole applies to all namespaces within a cluster.
#ClusterRole: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)
	rules?: [...#Rule] @protobuf(1,Rule,#"(gogoproto.jsontag)="rules""#,"(gogoproto.nullable)=false")

	// Metadata contains name, namespace, labels and annotations
	metadata?: #ObjectMeta @protobuf(3,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.nullable)=false")
}

// Role applies only to a single namespace.
#Role: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)
	rules?: [...#Rule] @protobuf(1,Rule,#"(gogoproto.jsontag)="rules""#,"(gogoproto.nullable)=false")

	// Metadata contains name, namespace, labels and annotations
	metadata?: #ObjectMeta @protobuf(4,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.nullable)=false")
}

// RoleRef maps groups to Roles or ClusterRoles.
#RoleRef: {
	// Type of role being referenced.
	type?: string @protobuf(1,string,#"(gogoproto.jsontag)="type""#)

	// Name of the resource being referenced
	name?: string @protobuf(2,string,#"(gogoproto.jsontag)="name""#)
}

#Subject: {
	// Type of object referenced (user or group)
	type?: string @protobuf(1,string,#"(gogoproto.jsontag)="type""#)

	// Name of the referenced object
	name?: string @protobuf(2,string,#"(gogoproto.jsontag)="name""#)
}

// ClusterRoleBinding grants the permissions defined in a ClusterRole referenced
// to a user or a set of users
#ClusterRoleBinding: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Subjects holds references to the objects the ClusterRole applies to
	subjects?: [...#Subject] @protobuf(1,Subject,#"(gogoproto.jsontag)="subjects""#,"(gogoproto.nullable)=false")

	// RoleRef references a ClusterRole in the current namespace
	roleRef?: #RoleRef @protobuf(2,RoleRef,name=role_ref,#"(gogoproto.jsontag)="role_ref""#,"(gogoproto.nullable)=false")

	// Metadata contains name, namespace, labels and annotations
	metadata?: #ObjectMeta @protobuf(4,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.nullable)=false")
}

// RoleBinding grants the permissions defined in a Role referenced to a user or
// a set of users
#RoleBinding: {
	@protobuf(option (gogoproto.face)=true)
	@protobuf(option (gogoproto.goproto_getters)=false)

	// Subjects holds references to the objects the Role applies to
	subjects?: [...#Subject] @protobuf(1,Subject,#"(gogoproto.jsontag)="subjects""#,"(gogoproto.nullable)=false")

	// RoleRef references a Role in the current namespace
	roleRef?: #RoleRef @protobuf(2,RoleRef,name=role_ref,#"(gogoproto.jsontag)="role_ref""#,"(gogoproto.nullable)=false")

	// Metadata contains name, namespace, labels and annotations
	metadata?: #ObjectMeta @protobuf(5,ObjectMeta,"(gogoproto.embed)",#"(gogoproto.jsontag)="metadata,omitempty""#,"(gogoproto.nullable)=false")
}
