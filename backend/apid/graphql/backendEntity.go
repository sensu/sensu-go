package graphql

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
)

type backendEntityImpl struct {
	svc ServiceConfig
}

func (b *backendEntityImpl) Meta(p graphql.ResolveParams) (interface{}, error) {
	v := p.Source.(interface{ GetMetadata() *corev2.ObjectMeta })
	return v.GetMetadata(), nil
}
