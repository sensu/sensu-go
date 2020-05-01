package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	assetsPathPrefix = "assets"
)

var (
	assetKeyBuilder = store.NewKeyBuilder(assetsPathPrefix)
)

func getAssetPath(asset *corev2.Asset) string {
	return assetKeyBuilder.WithResource(asset).Build(asset.Name)
}

// GetAssetsPath gets the path of the asset store.
func GetAssetsPath(ctx context.Context, name string) string {
	namespace := corev2.ContextNamespace(ctx)

	return assetKeyBuilder.WithNamespace(namespace).Build(name)
}

// DeleteAssetByName deletes an asset by name.
func (s *Store) DeleteAssetByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	err := Delete(ctx, s.client, GetAssetsPath(ctx, name))
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// GetAssets fetches all assets from the store
func (s *Store) GetAssets(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Asset, error) {
	assets := []*corev2.Asset{}
	err := List(ctx, s.client, GetAssetsPath, &assets, pred)
	return assets, err
}

// GetAssetByName gets an Asset by name.
func (s *Store) GetAssetByName(ctx context.Context, name string) (*corev2.Asset, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify namespace and name")}
	}

	var asset corev2.Asset
	if err := Get(ctx, s.client, GetAssetsPath(ctx, name), &asset); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	if asset.Labels == nil {
		asset.Labels = make(map[string]string)
	}
	if asset.Annotations == nil {
		asset.Annotations = make(map[string]string)
	}

	return &asset, nil
}

// UpdateAsset updates an asset.
func (s *Store) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	if err := asset.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getAssetPath(asset), asset.Namespace, asset)
}
