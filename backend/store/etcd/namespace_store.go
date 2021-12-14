package etcd

import (
	"context"
	"errors"
	"path"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	clientv3 "go.etcd.io/etcd/client/v3"
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
func (s *Store) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
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
			clientv3.Compare(clientv3.Version(namespaceKey), "=", 0)).
		Then(
			// Create it
			clientv3.OpPut(namespaceKey, string(namespaceBytes)),
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

	var getresp *clientv3.TxnResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		// Validate whether there are any resources referencing the namespace
		getresp, err = s.client.Txn(ctx).Then(
			clientv3.OpGet(checkKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(entityConfigKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(assetKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(handlerKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(mutatorKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(eventFilterKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(hookKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
			clientv3.OpGet(silencedKeyBuilder.WithNamespace(name).Build(), clientv3.WithPrefix(), clientv3.WithCountOnly()),
		).Commit()
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	for _, r := range getresp.Responses {
		if r.GetResponseRange().Count > 0 {
			return &store.ErrNotValid{Err: errors.New("namespace is not empty")}
		}
	}

	return Delete(ctx, s.client, getNamespacePath(name))
}

// GetNamespace returns a single namespace with the given name
func (s *Store) GetNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	var namespace corev2.Namespace
	err := Get(ctx, s.client, getNamespacePath(name), &namespace)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	return &namespace, nil
}

// ListNamespaces returns all namespaces
func (s *Store) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	namespaces := []*corev2.Namespace{}
	err := List(ctx, s.client, GetNamespacesPath, &namespaces, pred)
	return namespaces, err
}

// UpdateNamespace updates a namespace with the given object
func (s *Store) UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	bytes, err := proto.Marshal(namespace)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	return kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		_, err = s.client.Put(ctx, getNamespacePath(namespace.Name), string(bytes))
		return kvc.RetryRequest(n, err)
	})
}
