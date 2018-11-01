package graphql

import (
	"errors"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.ViewerFieldResolvers = (*viewerImpl)(nil)

//
// Implement QueryFieldResolvers
//

type viewerImpl struct {
	usersCtrl actions.UserController
	nsCtrl    actions.NamespacesController
}

func newViewerImpl(store store.Store) *viewerImpl {
	return &viewerImpl{
		usersCtrl: actions.NewUserController(store),
		nsCtrl:    actions.NewNamespacesController(store),
	}
}

// Namespaces implements response to request for 'namespaces' field.
func (r *viewerImpl) Namespaces(p graphql.ResolveParams) (interface{}, error) {
	return r.nsCtrl.Query(p.Context)
}

// User implements response to request for 'user' field.
func (r *viewerImpl) User(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	if claims := jwt.GetClaimsFromContext(ctx); claims != nil {
		return r.usersCtrl.Find(ctx, claims.Subject)
	}

	return nil, errors.New("user not found in context")
}
