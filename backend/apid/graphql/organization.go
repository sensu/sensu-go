package graphql

import (
	"errors"

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
func (r *orgImpl) ID(p graphql.ResolveParams) (string, error) {
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

// IconID implements response to request for 'iconId' field.
// Experimental. Value is not persisted in any way at this time and is simply
// derived from the name.
func (r *orgImpl) IconID(p graphql.ResolveParams) (schema.Icon, error) {
	org := p.Source.(*types.Organization)
	logger.WithField("name", org.Name).WithField("num", org.Name[0]%11).Info("finding icon")
	switch org.Name[0] % 11 {
	case 0:
		return schema.Icons.BRIEFCASE, nil
	case 1:
		return schema.Icons.DONUT, nil
	case 2:
		return schema.Icons.EMOTICON, nil
	case 3:
		return schema.Icons.ESPRESSO, nil
	case 4:
		return schema.Icons.EXPLORE, nil
	case 5:
		return schema.Icons.FIRE, nil
	case 6:
		return schema.Icons.HALFHEART, nil
	case 7:
		return schema.Icons.HEART, nil
	case 8:
		return schema.Icons.MUG, nil
	case 9:
		return schema.Icons.POLYGON, nil
	case 10:
		return schema.Icons.VISIBILITY, nil
	}
	return "", errors.New("exhausted list of icons")
}

// Environments implements response to request for 'environments' field.
func (r *orgImpl) Environments(p graphql.ResolveParams) (interface{}, error) {
	org := p.Source.(*types.Organization)
	return r.envCtrl.Query(p.Context, org.Name)
}
