package graphql

import (
	"time"

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
	factory ClientFactory
}

// Begin implements response to request for 'begin' field.
func (r *silencedImpl) Begin(p graphql.ResolveParams) (*time.Time, error) {
	s := p.Source.(*types.Silenced)
	return convertTs(s.Begin), nil
}

// Check implements response to request for 'check' field.
func (r *silencedImpl) Check(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.Silenced)
	ctx := contextWithNamespace(p.Context, src.Namespace)

	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchCheck(src.Check)
	return handleFetchResult(res, err)
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
