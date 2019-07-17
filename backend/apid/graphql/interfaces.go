package graphql

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type AssetClient interface {
	ListAssets(context.Context) ([]*corev2.Asset, error)
	FetchAsset(context.Context, string) (*corev2.Asset, error)
	CreateAsset(context.Context, *corev2.Asset) error
	UpdateAsset(context.Context, *corev2.Asset) error
}
