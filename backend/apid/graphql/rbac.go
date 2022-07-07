package graphql

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

//
// Implement ClusterRoleFieldResolvers
//

var _ schema.ClusterRoleFieldResolvers = (*clusterRoleImpl)(nil)

type clusterRoleImpl struct {
	schema.ClusterRoleAliases
}

// ID implements response to request for 'id' field.
func (*clusterRoleImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.ClusterRoleTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*clusterRoleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.ClusterRole)
	return ok
}

//
// Implement ClusterRoleBindingFieldResolvers
//

var _ schema.ClusterRoleBindingFieldResolvers = (*clusterRoleBindingImpl)(nil)

type clusterRoleBindingImpl struct {
	schema.ClusterRoleBindingAliases
}

// ID implements response to request for 'id' field.
func (*clusterRoleBindingImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.ClusterRoleBindingTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*clusterRoleBindingImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.ClusterRoleBinding)
	return ok
}

//
// Implement RoleFieldResolvers
//

var _ schema.RoleFieldResolvers = (*roleImpl)(nil)

type roleImpl struct {
	schema.RoleAliases
}

// ID implements response to request for 'id' field.
func (*roleImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.RoleTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*roleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Role)
	logger.Warnf("T: %T, OK: %s", s, ok)
	return ok
}

//
// Implement RoleBindingFieldResolvers
//

var _ schema.RoleBindingFieldResolvers = (*roleBindingImpl)(nil)

type roleBindingImpl struct {
	schema.RoleBindingAliases
}

// ID implements response to request for 'id' field.
func (*roleBindingImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.RoleBindingTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*roleBindingImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.RoleBinding)
	return ok
}
