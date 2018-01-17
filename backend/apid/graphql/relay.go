package graphql

import "github.com/sensu/sensu-go/backend/apid/graphql/schema"

var _ schema.PageInfoResolver = (*pageInfoImpl)(nil)

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
func (*entityImpl) HasNextPage(p graphql.ResolveParams) (bool, error) {
	i := p.Source.(pageInfo)
	return i.hasNextPage, nil
}

// HasPreviousPage implements response to request for 'hasPreviousPage' field.
func (*entityImpl) HasPreviousPage(p graphql.ResolveParams) (bool, error) {
	i := p.Source.(pageInfo)
	return i.hasPreviousPage, nil
}

// StartCursor implements response to request for 'startCursor' field.
func (*entityImpl) StartCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(pageInfo)
	return i.startCursor, nil
}

// EndCursor implements response to request for 'endCursor' field.
func (*entityImpl) EndCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(pageInfo)
	return i.endCursor, nil
}
