package etcd

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
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
		return err
	}

	b, err := proto.Marshal(ext)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(ext.Namespace)), ">", 0)
	req := clientv3.OpPut(getExtensionPath(ctx, ext.Name), string(b))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the extension %q in namespace %q",
			ext.Name, ext.Namespace,
		)
	}

	return nil
}

// DeregisterExtension deregisters an extension
func (s *Store) DeregisterExtension(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("no extension name specified")
	}

	_, err := s.client.Delete(ctx, getExtensionPath(ctx, name))
	return err
}

// GetExtension gets an extension
func (s *Store) GetExtension(ctx context.Context, name string) (*types.Extension, error) {
	if name == "" {
		return nil, errors.New("no extension name specified")
	}

	resp, err := s.client.Get(ctx, getExtensionPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, store.ErrNoExtension
	}

	var ext types.Extension
	return &ext, unmarshal(resp.Kvs[0].Value, &ext)
}

// GetExtensions gets an extension
func (s *Store) GetExtensions(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Extension, error) {
	extensions := []*types.Extension{}
	err := List(ctx, s.client, getExtensionPath, &extensions, pred)
	return extensions, err
}
