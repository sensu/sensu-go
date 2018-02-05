package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.PageInfoFieldResolvers = (*pageInfoImpl)(nil)

//
// Implement PageInfoFieldResolvers
//

type pageInfoImpl struct{}

// HasNextPage implements response to request for 'hasNextPage' field.
func (*pageInfoImpl) HasNextPage(p graphql.ResolveParams) (bool, error) {
	logger.WithField("params", p).Info("has next")
	i := p.Source.(relay.PageInfo)
	return i.HasNextPage, nil
}

// HasPreviousPage implements response to request for 'hasPreviousPage' field.
func (*pageInfoImpl) HasPreviousPage(p graphql.ResolveParams) (bool, error) {
	i := p.Source.(relay.PageInfo)
	return i.HasPreviousPage, nil
}

// StartCursor implements response to request for 'startCursor' field.
func (*pageInfoImpl) StartCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(relay.PageInfo)
	return i.StartCursor.String(), nil
}

// EndCursor implements response to request for 'endCursor' field.
func (*pageInfoImpl) EndCursor(p graphql.ResolveParams) (string, error) {
	i := p.Source.(relay.PageInfo)
	return i.EndCursor.String(), nil
}
