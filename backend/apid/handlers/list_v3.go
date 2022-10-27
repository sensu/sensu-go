package handlers

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func (h Handlers) ListResources(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	req := storev2.NewResourceRequestFromResource(h.Resource)
	req.Namespace = corev2.ContextNamespace(ctx)

	list, err := h.Store.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	return list.Unwrap()
}
