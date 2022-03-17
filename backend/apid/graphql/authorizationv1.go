package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
)

var _ schema.QueryExtensionSelfSubjectAccessReviewFieldResolvers = (*selfSubjectAccessReviewImpl)(nil)

type selfSubjectAccessReviewImpl struct {
	client GenericClient
}

// Cani implements response to request for 'cani' field.
func (r *selfSubjectAccessReviewImpl) Cani(p schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverParams) (interface{}, error) {
	if err := r.client.SetTypeMeta(*inputToTypeMeta(p.Args.Type)); err != nil {
		return util_api.HandleGetResult(nil, err)
	}
	ctx := contextWithNamespace(p.Context, p.Args.Meta.Namespace)
	err := r.client.Authorize(ctx, p.Args.Verb, p.Args.Meta.Name)
	return util_api.HandleGetResult(nil, err)
}
