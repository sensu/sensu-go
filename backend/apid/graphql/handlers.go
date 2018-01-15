package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.HandlerFieldResolvers = (*handlerImpl)(nil)
var _ schema.HandlerSocketFieldResolvers = (*handlerSocketImpl)(nil)

//
// Implement HandlerFieldResolvers
//

type handlerImpl struct {
	schema.HandlerAliases
	handlerController actions.HandlerController
	mutatorController actions.MutatorController
}

func newHandlerImpl(store store.Store) *handlerImpl {
	return &handlerImpl{
		handlerController: actions.NewHandlerController(store),
		mutatorController: actions.NewMutatorController(store),
	}
}

// ID implements response to request for 'id' field.
func (*handlerImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.HandlerTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (*handlerImpl) Namespace(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
}

// Mutator implements response to request for 'mutator' field.
func (r *handlerImpl) Mutator(p graphql.ResolveParams) (interface{}, error) {
	handler := p.Source.(*types.Handler)
	params := actions.QueryParams{"name": handler.Mutator}
	return r.mutatorController.Find(p.Context, params)
}

// Handlers implements response to request for 'handlers' field.
func (r *handlerImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
	handler := p.Source.(*types.Handler)
	if len(handler.Handlers) == 0 {
		return []interface{}{}, nil
	}

	handlers, err := r.handlerController.Query(p.Context)
	if err != nil {
		return nil, err
	}

	vals := make([]*types.Handler, 0, len(handler.Handlers))
	for _, handler := range handlers {
		if strings.InArray(handler.Name, handler.Handlers) {
			vals = append(vals, handler)
		}
		if len(handler.Handlers) == len(vals) {
			break
		}
	}
	return vals, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*handlerImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Entity)
	return ok
}

//
// Implement HandlerSocketFieldResolvers
//

type handlerSocketImpl struct{}

// Host implements response to request for 'host' field.
func (*handlerSocketImpl) Host(p graphql.ResolveParams) (string, error) {
	socket := p.Source.(*types.HandlerSocket)
	return socket.Host, nil
}

// Port implements response to request for 'port' field.
func (*handlerSocketImpl) Port(p graphql.ResolveParams) (int, error) {
	socket := p.Source.(*types.HandlerSocket)
	return int(socket.Port), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*handlerSocketImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.HandlerSocket)
	return ok
}
