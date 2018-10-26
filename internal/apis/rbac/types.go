package rbac

import "github.com/sensu/sensu-go/internal/apis/meta"

const (
	// APIGroupAll represents all possible API groups
	APIGroupAll = "*"
	// ResourceAll represents all possible resources
	ResourceAll = "*"
	// VerbAll represents all possible verbs
	VerbAll = "*"

	// GroupKind represents a group object in a subject
	GroupKind = "Group"
	// UserKind represents a user object in a subject
	UserKind = "User"
)

// A Rule holds information which describes an action that can be taken.
type Rule struct {
	// Verbs is a list of verbs that apply to all of the listed
	// resources for this rule. These include "get", "list", "watch",
	// "create", "update", "delete".
	// TODO: add support for "patch" (this is expensive and should be
	// delayed until a further release).
	// TODO: add support for "watch" (via websockets)
	Verbs []string `json:"verbs" protobuf:"bytes,1,rep,name=verbs"`

	// APIGroups is the name of the APIGroup that contains the resource
	APIGroups []string `json:"apiGroups" protobuf:"bytes,2,rep,name=apiGroups"`

	// Resources is a list of resources that this rule applies to.
	// "*" represents all resources.
	// TODO: enumerate "resources"
	Resources []string `json:"resources" protobuf:"bytes,3,rep,name=resources"`

	// ResourceNames is an optional list of resource names that the rule
	// applies to.
	ResourceNames []string `json:"resourceNames" protobuf:"bytes,4,rep,name=resourceNames"`
}

// A Role applies only to a single Namespace.
// +freeze-api:resource-name roles
type Role struct {
	meta.TypeMeta   `json:",inline" protobuf:"bytes,3,opt,name=typeMeta"`
	meta.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// Rules hold all of the Rules for this Role.
	Rules []Rule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// ClusterRole is a role that applies to all Namespaces within
// a cluster.
// +freeze-api:resource-name clusterRoles
type ClusterRole struct {
	meta.TypeMeta   `json:",inline" protobuf:"bytes,3,opt,name=typeMeta"`
	meta.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// Rules hold all of the Rules for this Role.
	Rules []Rule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
}

// RoleRef is used to map groups to Roles or ClusterRoles.
type RoleRef struct {
	// Type is the type of role being referenced.
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`

	// Name is the name of the resource being referenced.
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

// ClusterRoleBinding grants the permissions defined in a ClusterRole referenced
// to a user or a set of users
// +freeze-api:resource-name clusterRoleBindings
type ClusterRoleBinding struct {
	meta.TypeMeta   `json:",inline" protobuf:"bytes,4,opt,name=typeMeta"`
	meta.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// Subjects holds references to the objects the role applies to
	Subjects []Subject `json:"subjects" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef is the reference to a ClusterRole in the global namespace
	RoleRef RoleRef `json:"roleRef" protobuf:"bytes,3,name=roleRef"`
}

// RoleBinding grants the permissions defined in a Role referenced to a user or
// a set of users
// +freeze-api:resource-name roleBindings
type RoleBinding struct {
	meta.TypeMeta   `json:",inline" protobuf:"bytes,4,opt,name=typeMeta"`
	meta.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// Subjects holds references to the objects the role applies to
	Subjects []Subject `json:"subjects" protobuf:"bytes,2,rep,name=subjects"`

	// RoleRef is the reference to a Role in the current namespace
	RoleRef RoleRef `json:"roleRef" protobuf:"bytes,3,name=roleRef"`
}

// Subject contains a reference to the user identity a role binding applies to
// +freeze-api:resource-name subjects
type Subject struct {
	meta.TypeMeta `json:",inline" protobuf:"bytes,3,opt,name=typeMeta"`

	// Kind is the type of object referenced
	Kind string `json:"kind" protobuf:"bytes,1,name=kind"`

	// Name of the referenced object
	Name string `json:"name" protobuf:"bytes,2,name=name"`
}
