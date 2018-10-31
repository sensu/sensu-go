package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.HookFieldResolvers = (*hookImpl)(nil)
var _ schema.HookConfigFieldResolvers = (*hookCfgImpl)(nil)
var _ schema.HookListFieldResolvers = (*hookListImpl)(nil)

//
// Implement HookFieldResolvers
//

type hookImpl struct {
	schema.HookAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*hookImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Hook)
	return ok
}

//
// Implement HookConfigFieldResolvers
//

type hookCfgImpl struct {
	schema.HookConfigAliases
}

// ID implements response to request for 'id' field.
func (*hookCfgImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.HookTranslator.EncodeToString(p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*hookCfgImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.HookConfig)
	return ok
}

//
// Implement HookListFieldResolvers
//

type hookListImpl struct{}

// Hooks implements response to request for 'config' field.
func (*hookListImpl) Hooks(p graphql.ResolveParams) ([]string, error) {
	l, _ := p.Source.(types.HookList)
	return l.Hooks, nil
}

// Type implements response to request for 'type' field.
func (*hookListImpl) Type(p graphql.ResolveParams) (string, error) {
	l, _ := p.Source.(types.HookList)
	return l.Type, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*hookListImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.HookList)
	return ok
}
