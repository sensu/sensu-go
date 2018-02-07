package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.OrganizationFieldResolvers = (*orgImpl)(nil)

//
// Implement OrganizationFieldResolvers
//

type orgImpl struct {
	envCtrl actions.EnvironmentController
}

func newOrgImpl(store store.EnvironmentStore) *orgImpl {
	return &orgImpl{
		envCtrl: actions.NewEnvironmentController(store),
	}
}

// ID implements response to request for 'id' field.
func (r *orgImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.OrganizationTranslator.EncodeToString(p.Source), nil
}

// Name implements response to request for 'name' field.
func (r *orgImpl) Name(p graphql.ResolveParams) (string, error) {
	org := p.Source.(*types.Organization)
	return org.Name, nil
}

// Description implements response to request for 'description' field.
func (r *orgImpl) Description(p graphql.ResolveParams) (string, error) {
	org := p.Source.(*types.Organization)
	return org.Description, nil
}

// Environments implements response to request for 'environments' field.
func (r *orgImpl) Environments(p graphql.ResolveParams) (interface{}, error) {
	org := p.Source.(*types.Organization)
	return r.envCtrl.Query(p.Context, org.Name)
}
