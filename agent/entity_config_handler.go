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

	logger.Debug("received an entity config from the backend")

	// We rely on the special EntityNotFound entity name to determine whether
	// agentd found an entity, therefore only update the entity config if an
	// entity was found
	if entity.Metadata.Name != corev3.EntityNotFound {
		a.entityConfig = &entity
	}

	a.entityConfigCh <- struct{}

	return nil
}
