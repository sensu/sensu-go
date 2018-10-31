package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.RuleFieldResolvers = (*ruleImpl)(nil)
var _ schema.RoleFieldResolvers = (*roleImpl)(nil)

//
// Implement RuleFieldResolvers
//

type ruleImpl struct {
	schema.RuleAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*ruleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.Rule)
	return ok
}

//
// Implement RoleFieldResolvers
//

type roleImpl struct {
	schema.RoleAliases
}

// ID implements response to request for 'id' field.
func (*roleImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.RoleTranslator.EncodeToString(p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*roleImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Role)
	return ok
}
