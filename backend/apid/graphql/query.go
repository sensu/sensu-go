package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	store store.Store
}

// Checks implements response to request for 'checks' field.
func (r *queryImpl) Checks(p graphql.ResolveParams) (interface{}, error) {
	// TODO: Remove me
	checks, err := r.store.GetCheckConfigs(p.Context)
	return checks, err
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	// TODO: Re-implement lookup.
	return nil, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *queryImpl) IsTypeOf(_ interface{}, _ graphql.IsTypeOfParams) bool {
	return false
}

//
// Implement Node interface
//

type nodeImpl struct{}

func (nodeImpl) ResolveType(i interface{}, _ graphql.ResolveTypeParams) graphql.Type {
	// TODO: Re-implement lookup.
	return schema.CheckConfigType
}
