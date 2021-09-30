package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// ClusterRolesResource is the name of this resource type
	ClusterRolesResource = "clusterroles"

	// ClusterRoleBindingsResource is the name of this resource type
	ClusterRoleBindingsResource = "clusterrolebindings"

	// RolesResource is the name of this resource type
	RolesResource = "roles"

	// RoleBindingsResource is the name of this resource type
	RoleBindingsResource = "rolebindings"

	// ResourceAll represents all possible resources
	ResourceAll = "*"
	// VerbAll represents all possible verbs
	VerbAll = "*"

	// GroupType represents a group object in a subject
	GroupType = "Group"
	// UserType represents a user object in a subject
	UserType = "User"

	// ClusterRoleType represents a ClusterRole in a RoleRef
	ClusterRoleType = "ClusterRole"
	// RoleType represents a Role in a RoleRef
	RoleType = "Role"

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
	"events",
	"filters",
	"handlers",
	"hooks",
	"mutators",
	"silenced",
}

var allowedVerbs = []string{
	VerbAll,
	"get",
	"list",
	"create",
	"update",
	"delete",
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
		RoleRef:    FixtureRoleRef(RoleType, "read-write"),
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
		RoleRef:    FixtureRoleRef(ClusterRoleType, "read-write"),
	}
}

// StorePrefix returns the path prefix to this resource in the store
func (r *ClusterRole) StorePrefix() string {
	return "rbac/" + ClusterRolesResource
}

// URIPath returns the path component of a cluster role URI.
func (r *ClusterRole) URIPath() string {
	return path.Join(URLPrefix, ClusterRolesResource, url.PathEscape(r.Name))
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

	for i := range r.Rules {
		// Split the verbs, resources and resource names
		r.Rules[i].Verbs = split(r.Rules[i].Verbs)
		r.Rules[i].Resources = split(r.Rules[i].Resources)

		// Validate the verbs
		if err := validateVerbs(r.Rules[i].Verbs); err != nil {
			return err
		}
	}

	return nil
}

// StorePrefix returns the path prefix to this resource in the store
func (b *ClusterRoleBinding) StorePrefix() string {
	return "rbac/" + ClusterRoleBindingsResource
}

// URIPath returns the path component of a cluster role binding URI.
func (b *ClusterRoleBinding) URIPath() string {
	return path.Join(URLPrefix, ClusterRoleBindingsResource, url.PathEscape(b.Name))
}

// Validate a ClusterRoleBinding
func (b *ClusterRoleBinding) Validate() error {
	if err := ValidateSubscriptionName(b.Name); err != nil {
		return errors.New("the ClusterRoleBinding name " + err.Error())
	}

	if b.Namespace != "" {
		return errors.New("ClusterRoleBinding cannot have a namespace")
	}

	if err := ValidateRoleRef(&b.RoleRef); err != nil {
		return err
	}

	var err error
	b.Subjects, err = ValidateSubjects(b.Subjects)
	return err
}

// StorePrefix returns the path prefix to this resource in the store
func (r *Role) StorePrefix() string {
	return "rbac/" + RolesResource
}

// URIPath returns the path component of a role URI.
func (r *Role) URIPath() string {
	if r.Namespace == "" {
		return path.Join(URLPrefix, RolesResource, url.PathEscape(r.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(r.Namespace), RolesResource, url.PathEscape(r.Name))

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

	for i := range r.Rules {
		// Split the verbs, resources and resource names
		r.Rules[i].Verbs = split(r.Rules[i].Verbs)
		r.Rules[i].Resources = split(r.Rules[i].Resources)

		// Validate the verbs
		if err := validateVerbs(r.Rules[i].Verbs); err != nil {
			return err
		}
	}

	return nil
}

// StorePrefix returns the path prefix to this resource in the store
func (b *RoleBinding) StorePrefix() string {
	return "rbac/" + RoleBindingsResource
}

// URIPath returns the path component of a role binding URI.
func (b *RoleBinding) URIPath() string {
	if b.Namespace == "" {
		return path.Join(URLPrefix, RoleBindingsResource, url.PathEscape(b.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(b.Namespace), RoleBindingsResource, url.PathEscape(b.Name))
}

// Validate a RoleBinding
func (b *RoleBinding) Validate() error {
	if err := ValidateSubscriptionName(b.Name); err != nil {
		return errors.New("the RoleBinding name " + err.Error())
	}

	if b.Namespace == "" {
		return errors.New("the RoleBinding namespace must be set")
	}

	if err := ValidateRoleRef(&b.RoleRef); err != nil {
		return err
	}

	var err error
	b.Subjects, err = ValidateSubjects(b.Subjects)
	return err
}

// ValidateRoleRef checks that the role reference has a valid reference to
// either a Role or a ClusterRole
func ValidateRoleRef(roleRef *RoleRef) error {
	roleRef.Type = strings.Title(roleRef.Type)
	if roleRef.Type != ClusterRoleType && roleRef.Type != RoleType {
		return fmt.Errorf(
			"roleRef type %q is invalid, expected either %q or %q",
			roleRef.Type, ClusterRoleType, RoleType,
		)
	}

	if len(roleRef.Name) == 0 {
		return fmt.Errorf("roleRef name for %s is required", roleRef.Type)
	}

	return nil
}

// ValidateSubjects checks that there is at least one subject, and all subjects
// have non-empty types and names.
func ValidateSubjects(subjects []Subject) ([]Subject, error) {
	if len(subjects) == 0 {
		return subjects, errors.New("a RoleBinding must have at least one subject")
	}

	for i, subject := range subjects {
		subjects[i].Type = strings.Title(subject.Type)
		if subjects[i].Type != GroupType && subjects[i].Type != UserType {
			return subjects, fmt.Errorf(
				"subject type %q is invalid, expected either %q or %q",
				subject.Type, GroupType, UserType,
			)
		}
		if len(subject.Name) == 0 {
			return subjects, fmt.Errorf(
				"subject name for the %q type is required", subjects[i].Type,
			)
		}
	}

	return subjects, nil
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

// ClusterRoleFields returns a set of fields that represent that resource
func ClusterRoleFields(r Resource) map[string]string {
	resource := r.(*ClusterRole)
	fields := map[string]string{
		"clusterrole.name": resource.ObjectMeta.Name,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "clusterrole.labels.")
	return fields
}

// ClusterRoleBindingFields returns a set of fields that represent that resource
func ClusterRoleBindingFields(r Resource) map[string]string {
	resource := r.(*ClusterRoleBinding)
	fields := map[string]string{
		"clusterrolebinding.name":          resource.ObjectMeta.Name,
		"clusterrolebinding.role_ref.name": resource.RoleRef.Name,
		"clusterrolebinding.role_ref.type": resource.RoleRef.Type,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "clusterrolebinding.labels.")
	return fields
}

// RoleFields returns a set of fields that represent that resource
func RoleFields(r Resource) map[string]string {
	resource := r.(*Role)
	fields := map[string]string{
		"role.name":      resource.ObjectMeta.Name,
		"role.namespace": resource.ObjectMeta.Namespace,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "role.labels.")
	return fields
}

// RoleBindingFields returns a set of fields that represent that resource
func RoleBindingFields(r Resource) map[string]string {
	resource := r.(*RoleBinding)
	fields := map[string]string{
		"rolebinding.name":          resource.ObjectMeta.Name,
		"rolebinding.namespace":     resource.ObjectMeta.Namespace,
		"rolebinding.role_ref.name": resource.RoleRef.Name,
		"rolebinding.role_ref.type": resource.RoleRef.Type,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "rolebinding.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (r *ClusterRole) SetNamespace(namespace string) {
}

// SetObjectMeta sets the meta of the resource.
func (r *ClusterRole) SetObjectMeta(meta ObjectMeta) {
	r.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (b *ClusterRoleBinding) SetNamespace(namespace string) {
}

// SetObjectMeta sets the meta of the resource.
func (b *ClusterRoleBinding) SetObjectMeta(meta ObjectMeta) {
	b.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (r *Role) SetNamespace(namespace string) {
	r.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (r *Role) SetObjectMeta(meta ObjectMeta) {
	r.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (b *RoleBinding) SetNamespace(namespace string) {
	b.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (b *RoleBinding) SetObjectMeta(meta ObjectMeta) {
	b.ObjectMeta = meta
}

// RBACName returns the name of the resource for RBAC
func (*ClusterRoleBinding) RBACName() string {
	return "clusterrolebindings"
}

// RBACName returns the name of the resource for RBAC
func (*RoleBinding) RBACName() string {
	return "rolebindings"
}

// RBACName returns the name of the resource for RBAC
func (*ClusterRole) RBACName() string {
	return "clusterroles"
}

// RBACName returns the name of the resource for RBAC
func (*Role) RBACName() string {
	return "roles"
}

// split splits each string within a list using the comma seperator
func split(list []string) []string {
	var splitted []string

	for _, elem := range list {
		v := strings.Split(elem, ",")
		splitted = append(splitted, v...)
	}

	return splitted
}

// validateVerbs ensures the provided verbs are valid
func validateVerbs(verbs []string) error {
	for _, verb := range verbs {
		if !stringsutil.InArray(verb, allowedVerbs) {
			return fmt.Errorf("the verb %q is not valid", verb)
		}
	}

	return nil
}
