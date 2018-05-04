package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.ViewerFieldResolvers = (*viewerImpl)(nil)

//
// Implement QueryFieldResolvers
//

type viewerImpl struct {
	checksCtrl actions.CheckController
	entityCtrl actions.EntityController
	eventsCtrl actions.EventController
	usersCtrl  actions.UserController
	orgsCtrl   actions.OrganizationsController
}

func newViewerImpl(store store.Store, getter types.QueueGetter, bus messaging.MessageBus) *viewerImpl {
	return &viewerImpl{
		checksCtrl: actions.NewCheckController(store, getter),
		entityCtrl: actions.NewEntityController(store),
		eventsCtrl: actions.NewEventController(store, bus),
		usersCtrl:  actions.NewUserController(store),
		orgsCtrl:   actions.NewOrganizationsController(store),
	}
}

// Organizations implements response to request for 'organizations' field.
func (r *viewerImpl) Organizations(p graphql.ResolveParams) (interface{}, error) {
	return r.orgsCtrl.Query(p.Context)
}

// User implements response to request for 'user' field.
func (r *viewerImpl) User(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context
	actor := ctx.Value(types.AuthorizationActorKey).(authorization.Actor)
	return r.usersCtrl.Find(ctx, actor.Name)
}
