package agent

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

func (a *Agent) handleEntityConfig(ctx context.Context, payload []byte) error {
	var entity corev3.EntityConfig
	if err := a.unmarshal(payload, &entity); err != nil {
		return err
	}
	a.entityMu.Lock()
	defer a.entityMu.Unlock()
	a.entityConfig = &entity
	return nil
}
