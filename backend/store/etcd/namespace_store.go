package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/types"
)

const (
	namespacesPathPrefix = "namespaces"
)

func getNamespacePath(name string) string {
	return path.Join(EtcdRoot, namespacesPathPrefix, name)
}

// CreateNamespace creates a namespace with the provided namespace
func (s *Store) CreateNamespace(ctx context.Context, namespace *types.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return err
	}

	namespaceBytes, err := json.Marshal(namespace)
	if err != nil {
		return err
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
		return err
	}

	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the namespace %s",
			namespace.Name,
		)
	}

	return err
}

// DeleteNamespace deletes the namespace with the given name
func (s *Store) DeleteNamespace(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
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
			return errors.New("namespace is not empty") // TODO
		}
	}

	// Validate that there are no roles referencing the namespace
	roles, err := s.GetRoles(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, rule := range role.Rules {
			if rule.Namespace == name {
				return fmt.Errorf("namespace is not empty; role '%s' references it", role.Name)
			}
		}
	}

	// Delete the resource
	resp, err := s.client.Delete(ctx, getNamespacePath(name), v3.WithPrefix())
	if err != nil {
		return err
	}

	if resp.Deleted != 1 {
		return fmt.Errorf("namespace %s does not exist", name)
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
		return nil, err
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
func (s *Store) ListNamespaces(ctx context.Context) ([]*types.Namespace, error) {
	resp, err := s.client.Get(
		ctx,
		getNamespacePath(""),
		v3.WithPrefix(),
	)

	if err != nil {
		return []*types.Namespace{}, err
	}

	return unmarshalNamespaces(resp.Kvs)
}

// UpdateNamespace updates a namespace with the given object
func (s *Store) UpdateNamespace(ctx context.Context, namespace *types.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return err
	}

	bytes, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	_, err = s.client.Put(ctx, getNamespacePath(namespace.Name), string(bytes))

	return err
}

func unmarshalNamespaces(kvs []*mvccpb.KeyValue) ([]*types.Namespace, error) {
	s := make([]*types.Namespace, len(kvs))
	for i, kv := range kvs {
		namespace := &types.Namespace{}
		s[i] = namespace
		if err := json.Unmarshal(kv.Value, namespace); err != nil {
			return nil, err
		}
	}

	return s, nil
}
