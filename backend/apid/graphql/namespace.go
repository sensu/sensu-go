package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.NamespaceFieldResolvers = (*namespaceImpl)(nil)

type namespaceGetter interface {
	GetOrganization() string
	GetEnvironment() string
}

//
// Implement HandlerFieldResolvers
//

type namespaceImpl struct{}

// Organization implements response to request for 'organization' field.
func (*namespaceImpl) Organization(p graphql.ResolveParams) (string, error) {
	g := p.Source.(namespaceGetter)
	return g.GetOrganization(), nil
}

// Environment implements response to request for 'environment' field.
func (*namespaceImpl) Environment(p graphql.ResolveParams) (string, error) {
	g := p.Source.(namespaceGetter)
	return g.GetEnvironment(), nil
}
