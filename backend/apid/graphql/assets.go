package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

type assetImpl struct {
	schema.AssetAliases
}

// ID implements response to request for 'id' field.
func (r *assetImpl) ID(p graphql.ResolveParams) (interface{}, error) {
	return globalid.CheckTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (r *assetImpl) Namespace(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
}
