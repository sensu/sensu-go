package graphql

import (
	"sort"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

// KVPairString pair of values
type KVPairString struct {
	Key string
	Val string
}

// Labels implements response to request for 'labels' field.
func (r *objectMetaImpl) Labels(p graphql.ResolveParams) (interface{}, error) {
	return makeKVPairString(toObjectMeta(p.Source).Labels), nil
}

// Annotations implements response to request for 'annotations' field.
func (r *objectMetaImpl) Annotations(p graphql.ResolveParams) (interface{}, error) {
	return makeKVPairString(toObjectMeta(p.Source).Annotations), nil
}

func toObjectMeta(m interface{}) corev2.ObjectMeta {
	switch m := m.(type) {
	case corev2.ObjectMeta:
		return m
	case *corev2.ObjectMeta:
		return *m
	}
	return corev2.ObjectMeta{}
}

func makeKVPairString(m map[string]string) []KVPairString {
	pairs := make([]KVPairString, 0, len(m))
	for key, val := range m {
		pair := KVPairString{Key: key, Val: val}
		pairs = append(pairs, pair)
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})
	return pairs
}

func inputToTypeMeta(i *schema.TypeMetaInput) *corev2.TypeMeta {
	if i == nil {
		return nil
	}
	return &corev2.TypeMeta{
		Type:       i.Type,
		APIVersion: i.ApiVersion,
	}
}
