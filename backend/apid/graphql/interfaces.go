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

type CheckClient interface {
	CreateCheck(context.Context, *corev2.CheckConfig) error
	UpdateCheck(context.Context, *corev2.CheckConfig) error
	DeleteCheck(context.Context, string) error
	ExecuteCheck(context.Context, string, *corev2.AdhocRequest) error
	FetchCheck(context.Context, string) (*corev2.CheckConfig, error)
	ListChecks(context.Context) ([]*corev2.CheckConfig, error)
}
