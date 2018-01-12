package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

//
// Implement CheckConfigFieldResolvers
//

type checkCfgImpl struct {
	schema.CheckConfigAliases
	store interface {
		store.HandlerStore
	}
}

// ID implements response to request for 'id' field.
func (r *checkCfgImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.CheckTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (r *checkCfgImpl) Namespace(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
}

// Handlers implements response to request for 'handlers' field.
func (r *checkCfgImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
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
func (r *checkCfgImpl) IsTypeOf(p graphql.IsTypeOfParams) bool {
	_, ok := p.Value.(*types.CheckConfig)
	return ok
}

//
// Implement CheckFieldResolvers
//

type checkImpl struct {
	schema.CheckAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *checkImpl) IsTypeOf(p graphql.IsTypeOfParams) bool {
	_, ok := p.Value.(*types.Check)
	return ok
}

//
// Implement CheckHistoryFieldResolvers
//

type checkHistoryImpl struct {
	schema.CheckAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *checkHistoryImpl) IsTypeOf(p graphql.IsTypeOfParams) bool {
	_, ok := p.Value.(types.CheckHistory)
	return ok
}
