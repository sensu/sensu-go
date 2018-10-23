package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// HealthController exposes actions which a viewer can perform
type HealthController struct {
	store store.HealthStore
}

// NewHealthController returns new HealthController
func NewHealthController(store store.HealthStore) HealthController {
	return HealthController{
		store: store,
	}
}

// GetClusterHealth returns health information
func (h HealthController) GetClusterHealth(ctx context.Context) *types.HealthResponse {
	return h.store.GetClusterHealth(ctx)
}
