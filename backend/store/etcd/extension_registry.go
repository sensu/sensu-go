package etcd

import (
	"context"
	"errors"

	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"go.etcd.io/etcd/client/v3"
)

const (
	extRegistryPathPrefix = "extensions"
)

var (
	extKeyBuilder = store.NewKeyBuilder(extRegistryPathPrefix)
)

func getExtensionPath(ctx context.Context, name string) string {
	namespace := types.ContextNamespace(ctx)

	return extKeyBuilder.WithNamespace(namespace).Build(name)
}

// RegisterExtension registers an extension.
func (s *Store) RegisterExtension(ctx context.Context, ext *types.Extension) error {
	if err := ext.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	b, err := proto.Marshal(ext)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(ext.Namespace)), ">", 0)
	req := clientv3.OpPut(getExtensionPath(ctx, ext.Name), string(b))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	if !res.Succeeded {
		return &store.ErrNamespaceMissing{Namespace: ext.Namespace}
	}

	return nil
}

// DeregisterExtension deregisters an extension
func (s *Store) DeregisterExtension(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("no extension name specified")}
	}

	if _, err := s.client.Delete(ctx, getExtensionPath(ctx, name)); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	return nil
}

// GetExtension gets an extension
func (s *Store) GetExtension(ctx context.Context, name string) (*types.Extension, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("no extension name specified")}
	}

	resp, err := s.client.Get(ctx, getExtensionPath(ctx, name))
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	if len(resp.Kvs) == 0 {
		return nil, store.ErrNoExtension
	}

	var ext types.Extension
	if err := unmarshal(resp.Kvs[0].Value, &ext); err != nil {
		return nil, &store.ErrDecode{Err: err}
	}
	return &ext, nil
}

// GetExtensions gets an extension
func (s *Store) GetExtensions(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Extension, error) {
	extensions := []*types.Extension{}
	err := List(ctx, s.client, getExtensionPath, &extensions, pred)
	return extensions, err
}
