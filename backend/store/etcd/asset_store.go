package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	assetsPathPrefix = "assets"
)

func getAssetsPath(org, name string) string {
	return path.Join(etcdRoot, assetsPathPrefix, org, name)
}

// GetAssets fetches all assets from the store
func (s *etcdStore) GetAssets(org string) ([]*types.Asset, error) {
	// Verify that the organization exist
	if org != "" {
		if _, err := s.GetOrganizationByName(org); err != nil {
			return nil, fmt.Errorf("the organization '%s' is invalid", org)
		}
	}

	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(org, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetAssetByName(org, name string) (*types.Asset, error) {
	if org == "" || name == "" {
		return nil, errors.New("must specify organization and name")
	}

	// Verify that the organization exist
	if _, err := s.GetOrganizationByName(org); err != nil {
		return nil, fmt.Errorf("the organization '%s' is invalid", org)
	}

	resp, err := s.kvc.Get(context.TODO(), getAssetsPath(org, name))
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

	_, err = s.kvc.Put(context.TODO(), getAssetsPath(asset.Organization, asset.Name), string(assetBytes))
	if err != nil {
		return err
	}

	return nil
}

// TODO Cleanup associated checks?
func (s *etcdStore) DeleteAssetByName(org, name string) error {
	if org == "" || name == "" {
		return errors.New("must specify organization and name")
	}

	// Verify that the organization exist
	if _, err := s.GetOrganizationByName(org); err != nil {
		return fmt.Errorf("the organization '%s' is invalid", org)
	}

	_, err := s.kvc.Delete(context.TODO(), getAssetsPath(org, name))
	return err
}
