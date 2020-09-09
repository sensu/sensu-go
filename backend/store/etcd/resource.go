package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
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

func (s *Store) PatchResource(ctx context.Context, resource corev2.Resource, key string, patcher patch.Patcher, etag []byte) ([]byte, error) {
	// Get the stored resource along with the etcd response so we can use the
	// revision later to ensure the resource wasn't modified in the mean time
	resp, err := GetResponse(ctx, s.client, key, resource)
	if err != nil {
		return nil, err
	}
	version := resp.Kvs[0].Version

	// TODO(palourde): We should verify the etag here

	// Encode the stored resource
	original, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(original)
	if err != nil {
		return nil, err
	}

	// Decode the resulting document into provided resource
	if err := json.Unmarshal(patchedResource, &resource); err != nil {
		return nil, err
	}

	if err := UpdateWithVersion(ctx, s.client, key, resource, version); err != nil {
		return nil, err
	}

	// TODO(palourde): Calculate the new etag and return it here
	return nil, nil
}
