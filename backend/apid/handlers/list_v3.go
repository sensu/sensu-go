package handlers

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func (h Handlers) ListV3Resources(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	req := storev2.NewResourceRequest(ctx, corev2.ContextNamespace(ctx), "", h.V3Resource.StoreName())
	list, err := h.StoreV2.List(req, pred)
	if err != nil {
		return nil, err
	}
	return list.Unwrap()
}
