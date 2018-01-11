package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type checkImpl struct {
	schema.CheckAliases
	store interface {
		store.HandlerStore
	}
}

// ID implements response to request for 'id' field.
func (r *checkImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	// check := p.Source.(*types.CheckConfig)
	return nil
}

// Handlers implements response to request for 'handlers' field.
func (r *checkImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
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
func (r *checkIml) IsTypeOf(p graphql.IsTypeOfParams) bool {
	_, ok := p.Value.(*types.CheckConfig)
	return ok
}
