package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var _ schema.CheckFieldResolvers = (*checkImpl)(nil)
var _ schema.CheckConfigFieldResolvers = (*checkCfgImpl)(nil)
var _ schema.CheckHistoryFieldResolvers = (*checkHistoryImpl)(nil)

// QueueStore contains store and queue interfaces.
type QueueStore interface {
	store.Store
	queue.Get
}

//
// Implement CheckConfigFieldResolvers
//

type checkCfgImpl struct {
	schema.CheckConfigAliases
	handlerCtrl actions.HandlerController
}

func newCheckCfgImpl(store store.Store) *checkCfgImpl {
	return &checkCfgImpl{handlerCtrl: actions.NewHandlerController(store)}
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
	check := p.Source.(*types.CheckConfig)
	handlers, err := r.handlerCtrl.Query(p.Context)
	if err != nil {
		return nil, err
	}

	// Filter out irrevelant handlers
	for i := 0; i < len(handlers); {
		for _, h := range check.Handlers {
			if h == handlers[i].Name {
				continue
			}
		}
		handlers = append(handlers[:i], handlers[i+1:]...)
	}
	return handlers, nil
}

// IsTypeOf is used to determine if a given value is associated with the Check type
func (r *checkCfgImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.CheckConfig)
	return ok
}

//
// Implement CheckFieldResolvers
//

type checkImpl struct {
	schema.CheckAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *checkImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Check)
	return ok
}

//
// Implement CheckHistoryFieldResolvers
//

type checkHistoryImpl struct {
	schema.CheckHistoryAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *checkHistoryImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.CheckHistory)
	return ok
}
