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
	org := organization(ctx)

	return assetKeyBuilder.WithOrg(org).Build(name)
}

// TODO Cleanup associated checks?
func (s *etcdStore) DeleteAssetByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.kvc.Delete(ctx, getAssetsPath(ctx, name))
	return err
}

// GetAssets fetches all assets from the store
func (s *etcdStore) GetAssets(ctx context.Context) ([]*types.Asset, error) {
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
		assetArray[i] = asset
	}

	return assetArray, nil
}

func (s *etcdStore) GetAssetByName(ctx context.Context, name string) (*types.Asset, error) {
	if name == "" {
		return nil, errors.New("must specify organization and name")
	}

	resp, err := s.kvc.Get(ctx, getAssetsPath(ctx, name))
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

	cmp := clientv3.Compare(clientv3.Version(getOrganizationsPath(asset.Organization)), ">", 0)
	req := clientv3.OpPut(getAssetPath(asset), string(assetBytes))
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the asset %s in organization %s",
			asset.Name,
			asset.Organization,
		)
	}

	return nil
}
