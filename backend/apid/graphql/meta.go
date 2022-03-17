package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.ObjectMetaFieldResolvers = (*objectMetaImpl)(nil)

//
// Implement ObjectMetaFieldResolvers
//

type objectMetaImpl struct {
	schema.ObjectMetaAliases
}

// Labels implements response to request for 'labels' field.
func (r *objectMetaImpl) Labels(p graphql.ResolveParams) (interface{}, error) {
	return util_api.MakeKVPairString(util_api.ToObjectMeta(p.Source).Labels), nil
}

// Annotations implements response to request for 'annotations' field.
func (r *objectMetaImpl) Annotations(p graphql.ResolveParams) (interface{}, error) {
	return util_api.MakeKVPairString(util_api.ToObjectMeta(p.Source).Annotations), nil
}
