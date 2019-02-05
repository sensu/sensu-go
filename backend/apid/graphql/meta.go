package graphql

import (
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
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
	src := p.Source.(v2.ObjectMeta)
	return makeKVPairString(src.Labels), nil
}

// Annotations implements response to request for 'annotations' field.
func (r *objectMetaImpl) Annotations(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(v2.ObjectMeta)
	return makeKVPairString(src.Annotations), nil
}

func makeKVPairString(m map[string]string) []map[string]string {
	pairs := make([]map[string]string, 0, len(m))
	for key, val := range m {
		pair := make(map[string]string, 2)
		pair["key"] = key
		pair["val"] = val
		pairs = append(pairs, pair)
	}
	return pairs
}
