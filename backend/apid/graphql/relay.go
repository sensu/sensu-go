package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.PageInfoFieldResolvers = (*pageInfoImpl)(nil)

type pageInfo struct {
	hasNextPage     bool
	hasPreviousPage bool
	startCursor     string
	endCursor       string
}

//
// Implement PageInfoFieldResolvers
//

type pageInfoImpl struct {
	noLookup
}

// HasNextPage implements response to request for 'hasNextPage' field.
func (*pageInfoImpl) HasNextPage(p graphql.ResolveParams) (bool, error) {
	i := p.Source.(pageInfo)
	return i.hasNextPage, nil
}

// HasPreviousPage implements response to request for 'hasPreviousPage' field.
func (*pageInfoImpl) HasPreviousPage(p graphql.ResolveParams) (bool, error) {
	i := p.Source.(pageInfo)
	return i.hasPreviousPage, nil
}

// StartCursor implements response to request for 'startCursor' field.
func (*pageInfoImpl) StartCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(pageInfo)
	return i.startCursor, nil
}

// EndCursor implements response to request for 'endCursor' field.
func (*pageInfoImpl) EndCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(pageInfo)
	return i.endCursor, nil
}
