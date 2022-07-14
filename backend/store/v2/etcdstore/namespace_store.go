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
	namespaceStoreName = new(corev3.Namespace).StoreName()
)

type NamespaceStore struct {
	client *clientv3.Client
}

func NewNamespaceStore(client *clientv3.Client) *NamespaceStore {
	return &NamespaceStore{
		client: client,
	}
}

// Create creates a namespace using the given namespace struct.
func (s *NamespaceStore) CreateIfNotExists(ctx context.Context, namespace *corev3.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(namespace.StoreName()).Build(namespace.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.KeyIsNotFound(key),
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, namespace, key, comparator)
}

// CreateOrUpdate creates a namespace or updates it if it already exists.
func (s *NamespaceStore) CreateOrUpdate(ctx context.Context, namespace *corev3.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.NewKeyBuilder(namespace.StoreName()).Build(namespace.Metadata.Name)

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(""),
	)

	return s.create(ctx, namespace, key, comparator)
}

// Delete deletes a namespace using the given namespace name.
func (s *NamespaceStore) Delete(ctx context.Context, name string) error {
	empty, err := s.IsEmpty(ctx, name)
	if err != nil {
		return err
	}
	if !empty {
		return &store.ErrNamespaceNotEmpty{Namespace: name}
	}

	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", name, namespaceStoreName)
	return stor.Delete(ctx, req)
}

// Exists determines if a namespace exists.
func (s *NamespaceStore) Exists(ctx context.Context, name string) (bool, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", name, namespaceStoreName)
	return stor.Exists(ctx, req)
}

// Get retrieves a namespace by a given name.
func (s *NamespaceStore) Get(ctx context.Context, name string) (*corev3.Namespace, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", name, namespaceStoreName)
	wrapper, err := stor.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	namespace := &corev3.Namespace{}
	if err := wrapper.UnwrapInto(namespace); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return namespace, nil
}

// List returns all namespaces. A nil slice with no error is returned if none
// were found.
func (s *NamespaceStore) List(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", "", namespaceStoreName)
	wrapList, err := stor.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	namespaces := []*corev3.Namespace{}
	if err := wrapList.UnwrapInto(namespaces); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return namespaces, nil
}

func (s *NamespaceStore) Patch(ctx context.Context, name string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	// Fetch the current namespace & wrap it
	namespace, err := s.Get(ctx, name)
	if err != nil {
		return err
	}
	wrapper, err := storev2.WrapResource(namespace)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	// Patch the namespace
	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", name, namespaceStoreName)
	if err := stor.Patch(ctx, req, wrapper, patcher, conditions); err != nil {
		return err
	}
	return nil
}

// UpdateIfExists updates a given namespace.
func (s *NamespaceStore) UpdateIfExists(ctx context.Context, namespace *corev3.Namespace) error {
	wrapper, err := storev2.WrapResource(namespace)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	stor := NewStore(s.client)
	typeMeta := corev2.TypeMeta{Type: "Namespace", APIVersion: "core/v3"}
	req := storev2.NewResourceRequest(typeMeta, "", namespace.Metadata.Name, namespaceStoreName)
	if err := stor.UpdateIfExists(ctx, req, wrapper); err != nil {
		return err
	}
	return nil
}

func (s *NamespaceStore) IsEmpty(ctx context.Context, name string) (bool, error) {
	key := store.NewKeyBuilder(namespaceIndexStoreName).Build(name)
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithCountOnly())
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return resp.Count == 0, nil
}

func (s *NamespaceStore) create(ctx context.Context, namespace *corev3.Namespace, key string, comparator *kvc.Comparator) error {
	wrapped, err := wrap.Resource(namespace)
	if err != nil {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore could not wrap namespace resource: %v", err)}
	}

	msg, err := proto.Marshal(wrapped)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}

	op := clientv3.OpPut(key, string(msg))

	return kvc.Txn(ctx, s.client, comparator, op)
}
