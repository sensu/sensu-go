package relay

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/graphql"
)

// A Resolver retreives nodes and info using a given register.
type Resolver struct {
	Register *NodeRegister
}

// FindType finds the GraphQL type for the given resource.
func (r *Resolver) FindType(ctx context.Context, i interface{}) *graphql.Type {
	translator, err := globalid.ReverseLookup(i)
	if err != nil {
		return nil
	}

	components := translator.Encode(ctx, i)
	resolver := r.Register.Lookup(components)
	if resolver == nil {
		logger := logger.WithField("translator", fmt.Sprintf("%#v", translator))
		logger.Error("unable to find node resolver for type")
		return nil
	}
	return &resolver.ObjectType
}

// Find lookups and retrieves the resource associated with the given GlobalID.
func (r *Resolver) Find(ctx context.Context, id string, info graphql.ResolveInfo) (interface{}, error) {
	// Decode given ID
	idComponents, err := globalid.Decode(id)
	if err != nil {
		return nil, err
	}

	// Lookup resolver using components of a global ID
	resolver := r.Register.Lookup(idComponents)
	if resolver == nil {
		return nil, errors.New("unable to find type associated with this ID")
	}

	// Lift org & env into context
	ctx = context.WithValue(ctx, corev2.NamespaceKey, idComponents.Namespace())

	// Fetch resource from using resolver
	params := NodeResolverParams{
		Context:      ctx,
		IDComponents: idComponents,
		Info:         info,
	}
	return resolver.Resolve(params)
}
