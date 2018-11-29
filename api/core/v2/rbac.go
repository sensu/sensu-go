package v2

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	// ResourceAll represents all possible resources
	ResourceAll = "*"
	// VerbAll represents all possible verbs
	VerbAll = "*"

	// GroupType represents a group object in a subject
	GroupType = "Group"
	// UserType represents a user object in a subject
	UserType = "User"

	// LocalSelfUserResource represents a local user trying to view itself
	// or change its password
	LocalSelfUserResource = "localselfuser"
)

// CommonCoreResources represents the common "core" resources found in a
// namespace
var CommonCoreResources = []string{
	"assets",
	"checks",
	"entities",
	"extensions",
	"events",
	"filters",
	"handlers",
	"hooks",
	"mutators",
	"silenced",
}

// FixtureSubject creates a Subject for testing
func FixtureSubject(subjectType, name string) Subject {
	return Subject{
		Type: subjectType,
		Name: name,
	}
}

// FixtureRule returns a partial rule
func FixtureRule() Rule {
	return Rule{
		Verbs:     []string{VerbAll},
		Resources: []string{ResourceAll},
	}
}

// FixtureRole returns a partial role
func FixtureRole(name, namespace string) *Role {
	return &Role{
		ObjectMeta: NewObjectMeta(name, namespace),
		Rules: []Rule{
			FixtureRule(),
		},
	}
}

// FixtureRoleRef creates a RoleRef for testing
func FixtureRoleRef(roleType, name string) RoleRef {
	return RoleRef{
		Type: roleType,
		Name: name,
	}
}

// FixtureRoleBinding creates a RoleBinding for testing
func FixtureRoleBinding(name, namespace string) *RoleBinding {
	return &RoleBinding{
		ObjectMeta: NewObjectMeta(name, namespace),
		Subjects:   []Subject{FixtureSubject(UserType, "username")},
		RoleRef:    FixtureRoleRef("Role", "read-write"),
	}
}

// FixtureClusterRole returns a partial role
func FixtureClusterRole(name string) *ClusterRole {
	return &ClusterRole{
		ObjectMeta: NewObjectMeta(name, ""),
		Rules: []Rule{
			FixtureRule(),
		},
	}
}

// FixtureClusterRoleBinding creates a ClusterRoleBinding for testing
func FixtureClusterRoleBinding(name string) *ClusterRoleBinding {
	return &ClusterRoleBinding{
		ObjectMeta: NewObjectMeta(name, ""),
		Subjects:   []Subject{FixtureSubject(UserType, "username")},
		RoleRef:    FixtureRoleRef("ClusterRole", "read-write"),
	}
}

// Validate a ClusterRole
func (r *ClusterRole) Validate() error {
	if err := ValidateSubscriptionName(r.Name); err != nil {
		return errors.New("the ClusterRole name " + err.Error())
	}

	if len(r.Rules) == 0 {
		return errors.New("a ClusterRole must have at least one rule")
	}

	if r.Namespace != "" {
		return errors.New("ClusterRole cannot have a namespace")
	}

	return nil
}

// URIPath returns the path component of a ClusterRole URI.
func (r *ClusterRole) URIPath() string {
	return fmt.Sprintf("/api/core/v2/clusterroles/%s", url.PathEscape(r.Name))
}

// Validate a ClusterRoleBinding
func (b *ClusterRoleBinding) Validate() error {
	if err := ValidateSubscriptionName(b.Name); err != nil {
		return errors.New("the ClusterRoleBinding name " + err.Error())
	}

	if b.RoleRef.Name == "" || b.RoleRef.Type == "" {
		return errors.New("a ClusterRoleBinding needs a roleRef")
	}

	if len(b.Subjects) == 0 {
		return errors.New("a ClusterRoleBinding must have at least one subject")
	}

	if b.Namespace != "" {
		return errors.New("ClusterRoleBinding cannot have a namespace")
	}

	return nil
}

// URIPath returns the path component of a ClusterRole URI.
func (b *ClusterRoleBinding) URIPath() string {
	return fmt.Sprintf("/api/core/v2/clusterrolebindings/%s", url.PathEscape(b.Name))
}

// Validate a Role
func (r *Role) Validate() error {
	if err := ValidateSubscriptionName(r.Name); err != nil {
		return errors.New("the Role name " + err.Error())
	}

	if r.Namespace == "" {
		return errors.New("the Role namespace must be set")
	}

	if len(r.Rules) == 0 {
		return errors.New("a Role must have at least one rule")
	}

	return nil
}

// URIPath returns the path component of a Role URI.
func (r *Role) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/roles/%s",
		url.PathEscape(r.Namespace),
		url.PathEscape(r.Name),
	)
}

// Validate a RoleBinding
func (b *RoleBinding) Validate() error {
	if err := ValidateSubscriptionName(b.Name); err != nil {
		return errors.New("the RoleBinding name " + err.Error())
	}

	if b.Namespace == "" {
		return errors.New("the RoleBinding namespace must be set")
	}

	if b.RoleRef.Name == "" || b.RoleRef.Type == "" {
		return errors.New("a RoleBinding needs a roleRef")
	}

	if len(b.Subjects) == 0 {
		return errors.New("a RoleBinding must have at least one subject")
	}

	return nil
}

// URIPath returns the path component of a Role URI.
func (b *RoleBinding) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/rolebindings/%s",
		url.PathEscape(b.Namespace),
		url.PathEscape(b.Name),
	)
}

// ResourceMatches returns whether the specified requestedResource matches any
// of the rule resources
func (r Rule) ResourceMatches(requestedResource string) bool {
	for _, resource := range r.Resources {
		if resource == ResourceAll {
			return true
		}

		if resource == requestedResource {
			return true
		}
	}

	return false
}

// ResourceNameMatches returns whether the specified requestedResourceName
// matches any of the rule resources
func (r Rule) ResourceNameMatches(requestedResourceName string) bool {
	if len(r.ResourceNames) == 0 {
		return true
	}

	for _, name := range r.ResourceNames {
		if name == requestedResourceName {
			return true
		}
	}

	return false
}

// VerbMatches returns whether the specified requestedVerb matches any of the
// rule verbs
func (r Rule) VerbMatches(requestedVerb string) bool {
	for _, verb := range r.Verbs {
		if verb == VerbAll {
			return true
		}

		if verb == requestedVerb {
			return true
		}
	}

	return false
}

// NewRole creates a new Role.
func NewRole(meta ObjectMeta) *Role {
	return &Role{ObjectMeta: meta}
}

// NewRoleBinding creates a new RoleBinding.
func NewRoleBinding(meta ObjectMeta) *RoleBinding {
	return &RoleBinding{ObjectMeta: meta}
}
