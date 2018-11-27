package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
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

func getAssetsPath(ctx context.Context, name string) string {
	namespace := types.ContextNamespace(ctx)

	return assetKeyBuilder.WithNamespace(namespace).Build(name)
}

// DeleteAssetByName deletes an asset by name.
func (s *Store) DeleteAssetByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.client.Delete(ctx, getAssetsPath(ctx, name))
	return err
}

// GetAssets fetches all assets from the store
func (s *Store) GetAssets(ctx context.Context) ([]*types.Asset, error) {
	resp, err := query(ctx, s, getAssetsPath)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	assetArray := make([]*types.Asset, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		asset := &types.Asset{}
		err = json.Unmarshal(kv.Value, asset)
		if err != nil {
			return nil, err
		}
		if asset.Labels == nil {
			asset.Labels = make(map[string]string)
		}
		if asset.Annotations == nil {
			asset.Annotations = make(map[string]string)
		}
		assetArray[i] = asset
	}

	return assetArray, nil
}

// GetAssetByName gets an Asset by name.
func (s *Store) GetAssetByName(ctx context.Context, name string) (*types.Asset, error) {
	if name == "" {
		return nil, errors.New("must specify namespace and name")
	}

	resp, err := s.client.Get(ctx, getAssetsPath(ctx, name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	assetBytes := resp.Kvs[0].Value
	asset := &types.Asset{}
	if err := json.Unmarshal(assetBytes, asset); err != nil {
		return nil, err
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
		return err
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(asset.Namespace)), ">", 0)
	req := clientv3.OpPut(getAssetPath(asset), string(assetBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the asset %s in namespace %s",
			asset.Name,
			asset.Namespace,
		)
	}

	return nil
}
