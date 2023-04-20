package graphql

import (
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/core/v3/types"
)

type corev2PipelineImpl struct {
	schema.CoreV2PipelineAliases
}

// ID implements response to request for 'id' field.
func (*corev2PipelineImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.PipelineTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*corev2PipelineImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*corev2PipelineImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Pipeline)
	return ok
}
