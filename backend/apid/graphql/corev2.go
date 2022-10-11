package graphql

import (
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/types"
)

//
// CoreV2Pipeline
//

var _ schema.CoreV2PipelineExtensionOverridesFieldResolvers = (*corev2PipelineImpl)(nil)

type corev2PipelineImpl struct {
	schema.CoreV2PipelineAliases
}

// ID implements response to request for 'id' field.
func (*corev2PipelineImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.PipelineTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*corev2PipelineImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*corev2PipelineImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Pipeline)
	return ok
}

//
// Implement ClusterRoleFieldResolvers
//

var _ schema.CoreV2ClusterRoleExtensionOverridesFieldResolvers = (*clusterRoleImpl)(nil)

type clusterRoleImpl struct {
	schema.CoreV2ClusterRoleAliases
}

// ID implements response to request for 'id' field.
func (*clusterRoleImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.ClusterRoleTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*clusterRoleImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*clusterRoleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.ClusterRole)
	return ok
}

//
// Implement ClusterRoleBindingFieldResolvers
//

var _ schema.CoreV2ClusterRoleBindingExtensionOverridesFieldResolvers = (*clusterRoleBindingImpl)(nil)

type clusterRoleBindingImpl struct {
	schema.CoreV2ClusterRoleBindingAliases
}

// ID implements response to request for 'id' field.
func (*clusterRoleBindingImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.ClusterRoleBindingTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*clusterRoleBindingImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*clusterRoleBindingImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.ClusterRoleBinding)
	return ok
}

//
// Implement RoleFieldResolvers
//

var _ schema.CoreV2RoleExtensionOverridesFieldResolvers = (*roleImpl)(nil)

type roleImpl struct {
	schema.CoreV2RoleAliases
}

// ID implements response to request for 'id' field.
func (*roleImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.RoleTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*roleImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*roleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Role)
	return ok
}

//
// Implement RoleBindingFieldResolvers
//

var _ schema.CoreV2RoleBindingExtensionOverridesFieldResolvers = (*roleBindingImpl)(nil)

type roleBindingImpl struct {
	schema.CoreV2RoleBindingAliases
}

// ID implements response to request for 'id' field.
func (*roleBindingImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.RoleBindingTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*roleBindingImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*roleBindingImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.RoleBinding)
	return ok
}
