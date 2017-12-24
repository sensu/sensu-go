package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CheckResolver represents a collection of methods whose products represent the
// response values of the 'Check' type.
type CheckResolver struct {
	schema.CheckAliases
	store interface {
		store.HandlerStore
	}
}

// Handlers implements response to request for 'handlers' field.
func (r *CheckResolver) Handlers(p graphql.ResolveParams) interface{} {
	check := p.Source.(*types.CheckConfig)     // infer source
	handlers := r.store.GetHandlers(p.Context) // fetch all handlers

	// Filter out irrevelant handlers
	for i := 0; i < len(handlers); {
		for _, h := range check.Handlers {
			if h == handles[i].Name {
				continue
			}
		}
		handlers = append(a[:i], a[i+1:]...)
	}
	return handlers
}

// IsTypeOf is used to determine if a given value is associated with the Check type
func (r *CheckResolver) IsTypeOf(p graphql.IsTypeOfParams) bool {
	_, ok := p.Value.(*types.CheckConfig)
	return ok
}
