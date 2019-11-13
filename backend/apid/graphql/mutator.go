package graphql

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.MutatorFieldResolvers = (*mutatorImpl)(nil)

//
// Implement MutatorFieldResolvers
//

type mutatorImpl struct {
	schema.MutatorAliases
}

// ID implements response to request for 'id' field.
func (*mutatorImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.MutatorTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*mutatorImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Mutator)
	return ok
}

// ToJSON implements response to request for 'toJSON' field.
func (*mutatorImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}
