package agent

import (
	"time"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

func (a *Agent) getAgentEntity() *types.Entity {
	if a.entity == nil {
		meta := v2.NewObjectMeta(a.config.AgentName, a.config.Namespace)
		meta.Labels = a.config.Labels
		e := &types.Entity{
			EntityClass:   types.EntityAgentClass,
			Deregister:    a.config.Deregister,
			LastSeen:      time.Now().Unix(),
			Redact:        a.config.Redact,
			Subscriptions: a.config.Subscriptions,
			User:          a.config.User,
			ObjectMeta:    meta,
		}

		if a.config.DeregistrationHandler != "" {
			e.Deregistration = types.Deregistration{
				Handler: a.config.DeregistrationHandler,
			}
		}

		// Retrieve the system info from the agent's cached value
		a.systemInfoMu.RLock()
		e.System = *a.systemInfo
		a.systemInfoMu.RUnlock()

		a.entity = e
	}

	return a.entity
}

// getEntities receives an event and verifies if we have a proxy entity, so it
// can be added as the source, and ensures that the event uses the agent's
// entity
func (a *Agent) getEntities(event *types.Event) {
	// Verify if we have an entity in the event, and that it is different from the
	// agent's entity
	if event.Entity != nil && event.HasCheck() && event.Entity.Name != a.config.AgentName {
		// Identify the event's source as the provided entity so it can be properly
		// handled by the backend
		event.Check.ProxyEntityName = event.Entity.Name
	}

	// From this point we make sure that the agent's entity is used in the event
	// so we provide details like the namespace to the backend
	event.Entity = a.getAgentEntity()
}
