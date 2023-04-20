package graphql

import (
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/core/v3/types"
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

// Type implements response to request for 'Type' field.
func (*mutatorImpl) Type(p graphql.ResolveParams) (string, error) {
	m, ok := p.Source.(*corev2.Mutator)
	if !ok || m.Type == "" {
		return "pipe", nil
	}
	return m.Type, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*mutatorImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Mutator)
	return ok
}

// ToJSON implements response to request for 'toJSON' field.
func (*mutatorImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev3.Resource)), nil
}
