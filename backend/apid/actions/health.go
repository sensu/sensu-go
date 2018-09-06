package actions

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

type HealthController struct {
	store store.HealthStore
}

func NewHealthController(store store.HealthStore) HealthController {
	return HealthController{
		store: store,
	}
}

func (h HealthController) Health(ctx context.Context) *types.HealthResponse {
	return h.store.GetClusterHealth(ctx)
}
