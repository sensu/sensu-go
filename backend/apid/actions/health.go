package actions

import (
	corev2 "github.com/sensu/core/v2"
	"golang.org/x/net/context"
)

// HealthController exposes actions which a viewer can perform
type HealthController struct {
	// TODO decide if we want to remove this concept of health
}

// GetClusterHealth returns health information
func (h HealthController) GetClusterHealth(ctx context.Context) *corev2.HealthResponse {
	return &corev2.HealthResponse{}
}
