package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

	b, err := json.Marshal(ext)
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
	return &ext, json.Unmarshal(resp.Kvs[0].Value, &ext)
}

// GetExtensions gets an extension
func (s *Store) GetExtensions(ctx context.Context, pageSize int64, continueToken string) (extensions []*corev2.Extension, newContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getExtensionPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}

	if len(resp.Kvs) == 0 {
		return nil, "", nil
	}

	for _, kv := range resp.Kvs {
		var extension corev2.Extension
		if err := json.Unmarshal(kv.Value, &extension); err != nil {
			return nil, "", err
		}

		extensions = append(extensions, &extension)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastExtension := extensions[len(extensions)-1]
		newContinueToken = computeContinueToken(ctx, lastExtension)
	}

	return extensions, newContinueToken, nil
}
