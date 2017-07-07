package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	assetsPathPrefix = "assets"
)

func getAssetPath(asset *types.Asset) string {
	return path.Join(etcdRoot, assetsPathPrefix, asset.Organization, asset.Name)
}

func getAssetsPath(ctx context.Context, name string) string {
	var org string

	// Determine the organization
	if value := ctx.Value(types.OrganizationKey); value != nil {
		org = value.(string)
	} else {
		org = ""
	}

	return path.Join(etcdRoot, assetsPathPrefix, org, name)
}

// TODO Cleanup associated checks?
func (s *etcdStore) DeleteAssetByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.kvc.Delete(context.TODO(), getAssetsPath(ctx, name))
	return err
}

// GetAssets fetches all assets from the store
func (s *etcdStore) GetAssets(ctx context.Context) ([]*types.Asset, error) {
	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(ctx, ""), clientv3.WithPrefix())
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
		assetArray[i] = asset
	}

	return assetArray, nil
}

func (s *etcdStore) GetAssetByName(ctx context.Context, name string) (*types.Asset, error) {
	if name == "" {
		return nil, errors.New("must specify organization and name")
	}

	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(ctx, name))
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

	return asset, nil
}

func (s *etcdStore) UpdateAsset(ctx context.Context, asset *types.Asset) error {
	if err := asset.Validate(); err != nil {
		return err
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getAssetPath(asset), string(assetBytes))
	if err != nil {
		return err
	}

	return nil
}
