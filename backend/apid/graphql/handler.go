package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.HandlerFieldResolvers = (*handlerImpl)(nil)
var _ schema.HandlerSocketFieldResolvers = (*handlerSocketImpl)(nil)

//
// Implement HandlerFieldResolvers
//

type handlerImpl struct {
	schema.HandlerAliases
	handlerCtrl actions.HandlerController
	mutatorCtrl actions.MutatorController
}

func newHandlerImpl(store store.Store) *handlerImpl {
	return &handlerImpl{
		handlerCtrl: actions.NewHandlerController(store),
		mutatorCtrl: actions.NewMutatorController(store),
	}
}

// ID implements response to request for 'id' field.
func (*handlerImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.HandlerTranslator.EncodeToString(p.Source), nil
}

// Mutator implements response to request for 'mutator' field.
func (r *handlerImpl) Mutator(p graphql.ResolveParams) (interface{}, error) {
	handler := p.Source.(*types.Handler)
	return r.mutatorCtrl.Find(p.Context, handler.Mutator)
}

// Handlers implements response to request for 'handlers' field.
func (r *handlerImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
	handler := p.Source.(*types.Handler)
	ctx := types.SetContextFromResource(p.Context, handler)
	handlers, err := fetchHandlersWithNames(ctx, r.handlerCtrl, handler.Handlers)
	return handlers, err
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*handlerImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Handler)
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
