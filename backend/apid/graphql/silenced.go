package graphql

import (
	"time"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.SilencedFieldResolvers = (*silencedImpl)(nil)

//
// Implement CheckConfigFieldResolvers
//

type silencedImpl struct {
	schema.SilencedAliases
	nsFinder    namespaceFinder
	userFinder  userFinder
	checkFinder checkFinder
}

func newSilencedImpl(store store.Store, queue types.QueueGetter) *silencedImpl {
	nsCtrl := actions.NewNamespacesController(store)
	userCtrl := actions.NewUserController(store)
	checkCtrl := actions.NewCheckController(store, queue)
	return &silencedImpl{
		nsFinder:    nsCtrl,
		userFinder:  userCtrl,
		checkFinder: checkCtrl,
	}
}

// Begin implements response to request for 'begin' field.
func (r *silencedImpl) Begin(p graphql.ResolveParams) (*time.Time, error) {
	s := p.Source.(*types.Silenced)
	return convertTs(s.Begin), nil
}

// Creator implements response to request for 'creator' field.
func (r *silencedImpl) Creator(p graphql.ResolveParams) (interface{}, error) {
	sil := p.Source.(*types.Silenced)
	return handleControllerResults(r.userFinder.Find(p.Context, sil.Creator))
}

// Check implements response to request for 'check' field.
func (r *silencedImpl) Check(p graphql.ResolveParams) (interface{}, error) {
	sil := p.Source.(*types.Silenced)
	ctx := types.SetContextFromResource(p.Context, sil)
	return handleControllerResults(r.checkFinder.Find(ctx, sil.Check))
}

// Expires implements response to request for 'expires' field.
func (r *silencedImpl) Expires(p graphql.ResolveParams) (*time.Time, error) {
	s := p.Source.(*types.Silenced)
	if s.Expire > 0 {
		return convertTs(s.Begin + s.Expire), nil
	}
	return nil, nil
}

// ID implements response to request for 'id' field.
func (r *silencedImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.SilenceTranslator.EncodeToString(p.Source), nil
}
