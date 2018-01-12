package graphql

import (
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
	checks := r.store.GetCheckConfigs(p.Context)
	return checks, nil
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p graphql.ResolveParams) (interface{}, error) {
	// TODO: Re-implement lookup.
	return nil, nil
}

//
// Implement Node interface
//

type nodeImpl struct {
}

func (nodeImpl) ResolveType(i interface{}, _ graphql.ResolveParams) graphql.Type {
	// TODO: Re-implement lookup.
	return schema.CheckConfigType
}
