package etcd

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	assetsPathPrefix = "assets"
)

var (
	assetKeyBuilder = store.NewKeyBuilder(assetsPathPrefix)
)

func getAssetPath(asset *types.Asset) string {
	return assetKeyBuilder.WithResource(asset).Build(asset.Name)
}

// GetAssetsPath gets the path of the asset store.
func GetAssetsPath(ctx context.Context, name string) string {
	namespace := types.ContextNamespace(ctx)

	return assetKeyBuilder.WithNamespace(namespace).Build(name)
}

// DeleteAssetByName deletes an asset by name.
func (s *Store) DeleteAssetByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	if _, err := s.client.Delete(ctx, GetAssetsPath(ctx, name)); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	return nil
}

// GetAssets fetches all assets from the store
func (s *Store) GetAssets(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Asset, error) {
	assets := []*types.Asset{}
	err := List(ctx, s.client, GetAssetsPath, &assets, pred)
	return assets, err
}

// GetAssetByName gets an Asset by name.
func (s *Store) GetAssetByName(ctx context.Context, name string) (*types.Asset, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify namespace and name")}
	}

	resp, err := s.client.Get(ctx, GetAssetsPath(ctx, name))
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	assetBytes := resp.Kvs[0].Value
	asset := &types.Asset{}
	if err := unmarshal(assetBytes, asset); err != nil {
		return nil, &store.ErrDecode{Err: err}
	}
	if asset.Labels == nil {
		asset.Labels = make(map[string]string)
	}
	if asset.Annotations == nil {
		asset.Annotations = make(map[string]string)
	}

	return asset, nil
}

// UpdateAsset updates an asset.
func (s *Store) UpdateAsset(ctx context.Context, asset *types.Asset) error {
	if err := asset.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	assetBytes, err := proto.Marshal(asset)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(asset.Namespace)), ">", 0)
	req := clientv3.OpPut(getAssetPath(asset), string(assetBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	if !res.Succeeded {
		return &store.ErrNamespaceMissing{Namespace: asset.Namespace}
	}

	return nil
}
