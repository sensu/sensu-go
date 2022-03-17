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
	tm := util_api.InputToTypeMeta(p.Args.Type)
	if tm == nil { // type meta input is marked as non-null, if nil something is very very wrong
		panic("type meta is nil")
	}
	if err := r.client.SetTypeMeta(*tm); err != nil {
		return util_api.HandleGetResult(nil, err)
	}
	ctx := contextWithNamespace(p.Context, p.Args.Meta.Namespace)
	err := r.client.Authorize(ctx, p.Args.Verb, p.Args.Meta.Name)
	return util_api.HandleGetResult(nil, err)
}
