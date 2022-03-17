package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
)

var _ schema.QueryExtensionSelfSubjectAccessReviewFieldResolvers = (*selfSubjectAccessReviewImpl)(nil)

type selfSubjectAccessReviewImpl struct {
	client GenericClient
}

// Cani implements response to request for 'cani' field.
func (r *selfSubjectAccessReviewImpl) Cani(p schema.QueryExtensionSelfSubjectAccessReviewCaniFieldResolverParams) (interface{}, error) {
	if err := r.client.SetTypeMeta(*inputToTypeMeta(p.Args.Type)); err != nil {
		// ...
	}
}
