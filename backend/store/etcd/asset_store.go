package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// Asset

func getAssetsPath(name string) string {
	return fmt.Sprintf("%s/assets/%s", etcdRoot, name)
}

// GetAssets fetches all assets from the store
func (s *etcdStore) GetAssets() ([]*types.Asset, error) {
	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(""), clientv3.WithPrefix())
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

func (s *etcdStore) GetAssetByName(name string) (*types.Asset, error) {
	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(name))
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

func (s *etcdStore) UpdateAsset(asset *types.Asset) error {
	if err := asset.Validate(); err != nil {
		return err
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getAssetsPath(asset.Name), string(assetBytes))
	if err != nil {
		return err
	}

	return nil
}

// TODO Cleanup associated checks?
func (s *etcdStore) DeleteAssetByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getAssetsPath(name))
	return err
}
