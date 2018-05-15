package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.NamespaceFieldResolvers = (*namespaceImpl)(nil)

type namespaceGetter interface {
	GetOrganization() string
	GetEnvironment() string
}

//
// Implement NamespaceFieldResolvers
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

//
// Implement InterfaceTypeResolver for EnvironmentNode
//

type envNodeImpl struct{}

func (envNodeImpl) ResolveType(node interface{}, p graphql.ResolveTypeParams) *graphql.Type {
	// TODO: Prefer use of IsTypeOf for resolving type.
	return nil
}

func findEnvironment(ctx context.Context, finder environmentFinder, res types.MultitenantResource) (interface{}, error) {
	env, err := finder.Find(ctx, res.GetOrganization(), res.GetEnvironment())
	return handleControllerResults(env, err)
}

func findOrganization(ctx context.Context, finder organizationFinder, res types.MultitenantResource) (interface{}, error) {
	org, err := finder.Find(ctx, res.GetOrganization())
	return handleControllerResults(org, err)
}
