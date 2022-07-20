package etcdstore

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	entityStateStoreName = new(corev3.EntityState).StoreName()
)

type EntityStateStore struct {
	client *clientv3.Client
}

func NewEntityStateStore(client *clientv3.Client) *EntityStateStore {
	return &EntityStateStore{
		client: client,
	}
}

// Create creates an entity state using the given entity state struct.
func (s *EntityStateStore) CreateIfNotExists(ctx context.Context, state *corev3.EntityState) error {
	if err := state.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(state.StoreName()).Build(state.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.KeyIsNotFound(key),
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, state, key, comparator)
}

// CreateOrUpdate creates an entity state or updates it if it already exists.
func (s *EntityStateStore) CreateOrUpdate(ctx context.Context, state *corev3.EntityState) error {
	if err := state.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(state.StoreName()).Build(state.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, state, key, comparator)
}

// Delete deletes an entity state using the given namespace & name.
func (s *EntityStateStore) Delete(ctx context.Context, namespace, name string) error {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityStateStoreName)
	return stor.Delete(ctx, req)
}

// Exists determines if an entity state exists.
func (s *EntityStateStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityStateStoreName)
	return stor.Exists(ctx, req)
}

// Get retrieves an entity state by a given namespace & name.
func (s *EntityStateStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityState, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityStateStoreName)
	wrapper, err := stor.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	state := &corev3.EntityState{}
	if err := wrapper.UnwrapInto(state); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return state, nil
}

// List returns all entity states. A nil slice with no error is returned if
// none were found.
func (s *EntityStateStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityState, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, "", entityStateStoreName)
	wrapList, err := stor.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	states := []*corev3.EntityState{}
	if err := wrapList.UnwrapInto(states); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return states, nil
}

func (s *EntityStateStore) Patch(ctx context.Context, namespace, name string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	// Fetch the current entity state & wrap it
	state, err := s.Get(ctx, namespace, name)
	if err != nil {
		return err
	}
	wrapper, err := storev2.WrapResource(state)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	// Patch the entity state
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityStateStoreName)
	if err := stor.Patch(ctx, req, wrapper, patcher, conditions); err != nil {
		return err
	}
	return nil
}

// UpdateIfExists updates a given entity state.
func (s *EntityStateStore) UpdateIfExists(ctx context.Context, state *corev3.EntityState) error {
	wrapper, err := storev2.WrapResource(state)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"}
	namespace := state.Metadata.Namespace
	name := state.Metadata.Name
	req := storev2.NewResourceRequest(typeMeta, namespace, name, entityStateStoreName)
	if err := stor.UpdateIfExists(ctx, req, wrapper); err != nil {
		return err
	}
	return nil
}

func (s *EntityStateStore) create(ctx context.Context, state *corev3.EntityState, key string, comparator *kvc.Comparator) error {
	wrapped, err := wrap.Resource(state)
	if err != nil {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore could not wrap entity state resource: %v", err)}
	}

	msg, err := proto.Marshal(wrapped)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}
