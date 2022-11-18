package etcdstore

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	entityConfigStoreName = new(corev3.EntityConfig).StoreName()
)

type EntityConfigStore struct {
	client *clientv3.Client
}

func NewEntityConfigStore(client *clientv3.Client) *EntityConfigStore {
	return &EntityConfigStore{
		client: client,
	}
}

// Create creates an entity config using the given entity config struct.
func (s *EntityConfigStore) CreateIfNotExists(ctx context.Context, config *corev3.EntityConfig) error {
	if err := config.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(config.StoreName()).
		WithNamespace(config.Metadata.Namespace).
		Build(config.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.KeyIsNotFound(key),
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, config, key, comparator)
}

// CreateOrUpdate creates an entity config or updates it if it already exists.
func (s *EntityConfigStore) CreateOrUpdate(ctx context.Context, config *corev3.EntityConfig) error {
	if err := config.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(config.StoreName()).Build(config.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, config, key, comparator)
}

// Delete deletes an entity config using the given namespace & name.
func (s *EntityConfigStore) Delete(ctx context.Context, namespace, name string) error {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityConfigStoreName)
	return stor.Delete(ctx, req)
}

// Exists determines if an entity config exists.
func (s *EntityConfigStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityConfigStoreName)
	return stor.Exists(ctx, req)
}

// Get retrieves an entity config by a given namespace & name.
func (s *EntityConfigStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityConfig, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityConfigStoreName)
	wrapper, err := stor.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	config := &corev3.EntityConfig{}
	if err := wrapper.UnwrapInto(config); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return config, nil
}

// List returns all entity configs. A nil slice with no error is returned if
// none were found.
func (s *EntityConfigStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, "", entityConfigStoreName)
	wrapList, err := stor.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	configs := []*corev3.EntityConfig{}
	if err := wrapList.UnwrapInto(configs); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return configs, nil
}
func (s *EntityConfigStore) Count(ctx context.Context, namespace, eClass string) (int, error) {
	return 0, nil
}

func (s *EntityConfigStore) Watch(context.Context, storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	return nil
}

func (s *EntityConfigStore) Patch(ctx context.Context, namespace, name string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	// Fetch the current entity config & wrap it
	config, err := s.Get(ctx, namespace, name)
	if err != nil {
		return err
	}
	wrapper, err := storev2.WrapResource(config)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	// Patch the entity config
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityConfigStoreName)
	if err := stor.Patch(ctx, req, wrapper, patcher, conditions); err != nil {
		return err
	}
	return nil
}

// UpdateIfExists updates a given entity config.
func (s *EntityConfigStore) UpdateIfExists(ctx context.Context, config *corev3.EntityConfig) error {
	wrapper, err := storev2.WrapResource(config)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"}
	namespace := config.Metadata.Namespace
	name := config.Metadata.Name
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityConfigStoreName)
	if err := stor.UpdateIfExists(ctx, req, wrapper); err != nil {
		return err
	}
	return nil
}

func (s *EntityConfigStore) create(ctx context.Context, config *corev3.EntityConfig, key string, comparator *kvc.Comparator) error {
	wrapped, err := wrap.Resource(config)
	if err != nil {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore could not wrap entity config resource: %v", err)}
	}

	msg, err := proto.Marshal(wrapped)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}
