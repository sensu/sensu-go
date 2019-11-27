package etcd

import (
	"context"
	"errors"
	"path"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	namespacesPathPrefix = "namespaces"
)

func getNamespacePath(name string) string {
	return path.Join(EtcdRoot, namespacesPathPrefix, name)
}

// GetNamespacesPath gets the path of the namespace store.
func GetNamespacesPath(ctx context.Context, name string) string {
	return path.Join(EtcdRoot, namespacesPathPrefix, name)
}

// CreateNamespace creates a namespace with the provided namespace
func (s *Store) CreateNamespace(ctx context.Context, namespace *types.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	namespaceBytes, err := proto.Marshal(namespace)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	namespaceKey := getNamespacePath(namespace.Name)

	res, err := s.client.Txn(ctx).
		If(
			// Ensure the namespace does not already exist
			v3.Compare(v3.Version(namespaceKey), "=", 0)).
		Then(
			// Create it
			v3.OpPut(namespaceKey, string(namespaceBytes)),
		).Commit()
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	if !res.Succeeded {
		return &store.ErrAlreadyExists{Key: namespaceKey}
	}

	return nil
}

// DeleteNamespace deletes the namespace with the given name
func (s *Store) DeleteNamespace(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	// Validate whether there are any resources referencing the namespace
	getresp, err := s.client.Txn(ctx).Then(
		v3.OpGet(checkKeyBuilder.WithNamespace(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(entityKeyBuilder.WithNamespace(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(assetKeyBuilder.WithNamespace(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(handlerKeyBuilder.WithNamespace(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
		v3.OpGet(mutatorKeyBuilder.WithNamespace(name).Build(), v3.WithPrefix(), v3.WithCountOnly()),
	).Commit()
	if err != nil {
		return err
	}
	for _, r := range getresp.Responses {
		if r.GetResponseRange().Count > 0 {
			return &store.ErrNotValid{Err: errors.New("namespace is not empty")}
		}
	}

	// Delete the resource
	resp, err := s.client.Delete(ctx, getNamespacePath(name), v3.WithPrefix())
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	if resp.Deleted != 1 {
		return &store.ErrNotFound{Key: getNamespacePath(name)}
	}

	return nil
}

// GetNamespace returns a single namespace with the given name
func (s *Store) GetNamespace(ctx context.Context, name string) (*types.Namespace, error) {
	resp, err := s.client.Get(
		ctx,
		getNamespacePath(name),
		v3.WithLimit(1),
	)
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	namespaces, err := unmarshalNamespaces(resp.Kvs)
	if err != nil {
		return &types.Namespace{}, err
	}

	return namespaces[0], nil
}

// ListNamespaces returns all namespaces
func (s *Store) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Namespace, error) {
	namespaces := []*types.Namespace{}
	err := List(ctx, s.client, GetNamespacesPath, &namespaces, pred)
	return namespaces, err
}

// UpdateNamespace updates a namespace with the given object
func (s *Store) UpdateNamespace(ctx context.Context, namespace *types.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	bytes, err := proto.Marshal(namespace)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	if _, err := s.client.Put(ctx, getNamespacePath(namespace.Name), string(bytes)); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func unmarshalNamespaces(kvs []*mvccpb.KeyValue) ([]*types.Namespace, error) {
	s := make([]*types.Namespace, len(kvs))
	for i, kv := range kvs {
		namespace := &types.Namespace{}
		s[i] = namespace
		if err := unmarshal(kv.Value, namespace); err != nil {
			return nil, &store.ErrDecode{Err: err}
		}
	}

	return s, nil
}
