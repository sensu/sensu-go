package etcd

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateResource creates the given resource only if it does not already exist
func (s *Store) CreateResource(ctx context.Context, resource corev2.Resource) error {
	if err := resource.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.KeyFromResource(resource)
	namespace := resource.GetObjectMeta().Namespace

	msg, ok := resource.(proto.Message)
	if !ok {
		return &store.ErrEncode{Key: key, Err: fmt.Errorf("%T is not proto.Message", resource)}
	}

	return Create(ctx, s.client, key, namespace, msg)
}

// CreateOrUpdateResource creates or updates the given resource regardless of
// whether it already exists or not
func (s *Store) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error {
	if err := resource.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.KeyFromResource(resource)
	namespace := resource.GetObjectMeta().Namespace
	return CreateOrUpdate(ctx, s.client, key, namespace, resource)
}

// DeleteResource deletes the resource using the given resource prefix and name
func (s *Store) DeleteResource(ctx context.Context, resourcePrefix, name string) error {
	key := store.KeyFromArgs(ctx, resourcePrefix, name)
	return Delete(ctx, s.client, key)
}

// GetResource retrieves a resource with the given name and stores it into the
// resource pointer
func (s *Store) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
	key := store.KeyFromArgs(ctx, resource.StorePrefix(), name)
	return Get(ctx, s.client, key, resource)
}

// ListResources retrieves all resources for the resourcePrefix type and stores
// them into the resources pointer
func (s *Store) ListResources(ctx context.Context, resourcePrefix string, resources interface{}, pred *store.SelectionPredicate) error {
	keyBuilderFunc := func(ctx context.Context, name string) string {
		return store.NewKeyBuilder(resourcePrefix).WithContext(ctx).Build("")
	}

	return List(ctx, s.client, keyBuilderFunc, resources, pred)
}
