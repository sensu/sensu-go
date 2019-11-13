package graphql

import (
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

type assetImpl struct {
	schema.AssetAliases
}

// ID implements response to request for 'id' field.
func (*assetImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.AssetTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*assetImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*v2.Asset)
	return ok
}

// ToJSON implements response to request for 'toJSON' field.
func (*assetImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(v2.Resource)), nil
}
