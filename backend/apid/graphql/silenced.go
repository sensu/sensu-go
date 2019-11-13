package graphql

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.SilencedFieldResolvers = (*silencedImpl)(nil)

//
// Implement CheckConfigFieldResolvers
//

type silencedImpl struct {
	schema.SilencedAliases
	client CheckClient
}

// Begin implements response to request for 'begin' field.
func (r *silencedImpl) Begin(p graphql.ResolveParams) (*time.Time, error) {
	s := p.Source.(*corev2.Silenced)
	return convertTs(s.Begin), nil
}

// Check implements response to request for 'check' field.
func (r *silencedImpl) Check(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*corev2.Silenced)
	ctx := contextWithNamespace(p.Context, src.Namespace)

	res, err := r.client.FetchCheck(ctx, src.Check)
	return handleFetchResult(res, err)
}

// Expires implements response to request for 'expires' field.
func (r *silencedImpl) Expires(p graphql.ResolveParams) (*time.Time, error) {
	s := p.Source.(*corev2.Silenced)
	if s.Expire > 0 {
		return convertTs(s.Begin + s.Expire), nil
	}
	return nil, nil
}

// ID implements response to request for 'id' field.
func (r *silencedImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.SilenceTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (r *silencedImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}
