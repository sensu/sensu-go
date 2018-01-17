package graphql

import (
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

// IsTypeOf is used to determine if a given value is associated with the type
func (*mutatorImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Mutator)
	return ok
}
