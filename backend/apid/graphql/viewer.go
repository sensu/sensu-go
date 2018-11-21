package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/graphql"
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
	client := r.factory.NewWithContext(p.Context)
	return fetchNamespaces(client, nil)
}

// User implements response to request for 'user' field.
func (r *viewerImpl) User(p graphql.ResolveParams) (interface{}, error) {
	claims := jwt.GetClaimsFromContext(p.Context)
	logger.WithField("claims", claims).Info("huh")
	if claims == nil {
		return nil, nil
	}

	client := r.factory.NewWithContext(p.Context)
	res, err := client.FetchUser(claims.Subject)
	return handleFetchResult(res, err)
}
