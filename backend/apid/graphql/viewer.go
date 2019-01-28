package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.ViewerFieldResolvers = (*viewerImpl)(nil)

//
// Implement QueryFieldResolvers
//

type viewerImpl struct {
	factory ClientFactory
}

// Namespaces implements response to request for 'namespaces' field.
func (r *viewerImpl) Namespaces(p graphql.ResolveParams) (interface{}, error) {
	results, err := loadNamespaces(p.Context)
	records := make([]*types.Namespace, len(results))
	for i := range results {
		records[i] = &results[i]
	}
	return records, err
}

// User implements response to request for 'user' field.
func (r *viewerImpl) User(p graphql.ResolveParams) (interface{}, error) {
	claims := jwt.GetClaimsFromContext(p.Context)
	if claims == nil {
		return nil, nil
	}

	client := r.factory.NewWithContext(p.Context)
	res, err := client.FetchUser(claims.Subject)
	return handleFetchResult(res, err)
}
