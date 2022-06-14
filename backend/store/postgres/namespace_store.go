package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type NamespaceStore struct {
	store *StoreV2
}

func NewNamespaceStore(db *pgxpool.Pool, client *clientv3.Client) *NamespaceStore {
	return &NamespaceStore{
		store: NewStoreV2(db, client),
	}
}

// CreateNamespace creates a namespace using the given namespace struct.
func (e *NamespaceStore) CreateNamespace(ctx context.Context, n *corev2.Namespace) error {
	namespace := corev3.V2NamespaceToV3(n)
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true
	wrappedNamespace, err := storev2.WrapResource(namespace)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	if err := e.store.CreateIfNotExists(req, wrappedNamespace); err != nil {
		return err
	}
	return nil
}

// DeleteNamespace deletes a namespace using the given namespace name.
func (e *NamespaceStore) DeleteNamespace(ctx context.Context, name string) error {
	namespace := corev3.NewNamespace(name)
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true

	if err := e.store.Delete(req); err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
	}
	return nil
}

// DeleteNamespaceByName deletes a namespace using the given name.
func (e *NamespaceStore) DeleteNamespaceByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	return e.DeleteNamespace(ctx, name)
}

// GetNamespaces returns all namespaces. A nil slice with no error is returned
// if none were found.
func (e *NamespaceStore) GetNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	// Fetch the namespaces with the selection predicate
	req := storev2.ResourceRequest{
		Context:     ctx,
		StoreName:   new(corev3.Namespace).StoreName(),
		UsePostgres: true,
	}
	if pred.Ordering == corev3.NamespaceSortName {
		req.SortOrder = storev2.SortAscend
		if pred.Descending {
			req.SortOrder = storev2.SortDescend
		}
	}

	wNamespaces, err := e.store.List(req, pred)
	if err != nil {
		return nil, err
	}

	namespaces := make([]*corev3.Namespace, wNamespaces.Len())
	if err := wNamespaces.UnwrapInto(&namespaces); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: etcdstore.StoreKey(req)}
	}

	return namespaces, nil
}

func (e *NamespaceStore) GetNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	namespace := corev3.NewNamespace(name)
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true
	wrapper, err := e.store.Get(req)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}

	if err := wrapper.UnwrapInto(namespace); err != nil {
		return nil, err
	}
	return corev3.V3NamespaceToV2(namespace), nil
}

// GetNamespaceByName returns a namespace using the given name. The resulting
// namespace config is nil if none was found.
func (e *NamespaceStore) GetNamespaceByName(ctx context.Context, name string) (*corev3.Namespace, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	namespace := &corev3.Namespace{
		Metadata: &corev2.ObjectMeta{
			Name: name,
		},
	}
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true
	wrapper, err := e.store.Get(req)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}

	if err := wrapper.UnwrapInto(namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}

func (e *NamespaceStore) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	return nil, nil
}

// UpdateNamespace creates or updates a given namespace.
func (e *NamespaceStore) UpdateNamespace(ctx context.Context, n *corev2.Namespace) error {
	namespace := corev3.V2NamespaceToV3(n)
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	req.UsePostgres = true
	wrappedNamespace, err := storev2.WrapResource(namespace)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	if err := e.store.CreateOrUpdate(req, wrappedNamespace); err != nil {
		return err
	}
	return nil
}
