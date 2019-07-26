package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

type EntityClient struct {
	store store.ResourceStore
	auth  authorization.Authorizer
}

func NewEntityClient(store store.ResourceStore, auth authorization.Authorizer) *EntityClient {
	return &EntityClient{store: store, auth: auth}
}

//func (e *EntityClient) DeleteEntity(ctx context.Context, name string) error {
//	attrs := entityDeleteAttributes(ctx, name)
//	return nil
//}

func entityDeleteAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "entities",
		Verb:         "delete",
		ResourceName: name,
	}
}
