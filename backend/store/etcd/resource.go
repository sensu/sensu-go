package etcd

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// CreateResource ...
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

// CreateOrUpdateResource ...
func (s *Store) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error {
	if err := resource.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.KeyFromResource(resource)
	namespace := resource.GetObjectMeta().Namespace
	return CreateOrUpdate(ctx, s.client, key, namespace, resource)
}

// DeleteResource ...
func (s *Store) DeleteResource(ctx context.Context, kind, name string) error {
	key := store.KeyFromArgs(ctx, kind, name)
	return Delete(ctx, s.client, key)
}

// GetResource ...
func (s *Store) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
	key := store.KeyFromArgs(ctx, resource.StorePath(), name)
	return Get(ctx, s.client, key, resource)
}

// ListResources ...
func (s *Store) ListResources(ctx context.Context, kind string, resources interface{}, pred *store.SelectionPredicate) error {
	keyBuilderFunc := func(ctx context.Context, name string) string {
		return store.NewKeyBuilder(kind).WithContext(ctx).Build("")
	}

	return List(ctx, s.client, keyBuilderFunc, resources, pred)
}
