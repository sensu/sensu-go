package agent

import (
	"encoding/json"

	"github.com/sensu/sensu-go/system"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

func (a *Agent) getAgentEntity() *types.Entity {
	if a.entity == nil {
		e := &types.Entity{
			Class:            types.EntityAgentClass,
			Deregister:       a.config.Deregister,
			Environment:      a.config.Environment,
			ID:               a.config.AgentID,
			KeepaliveTimeout: a.config.KeepaliveTimeout,
			Organization:     a.config.Organization,
			Redact:           a.config.Redact,
			Subscriptions:    a.config.Subscriptions,
			User:             a.config.User,
		}

		if a.config.DeregistrationHandler != "" {
			e.Deregistration = types.Deregistration{
				Handler: a.config.DeregistrationHandler,
			}
		}

		// Set any extended attributes in the entity
		var attrMap map[string]interface{}
		err := json.Unmarshal(a.config.ExtendedAttributes, &attrMap)
		if err != nil {
			logger.WithError(err)
		}
		for k, v := range attrMap {
			err = dynamic.SetField(e, k, v)
			if err != nil {
				logger.WithError(err)
			}
		}

		s, err := system.Info()
		if err == nil {
			e.System = s
		}

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
	if event.Entity != nil && event.Entity.ID != a.config.AgentID {
		// Identify the event's source as the provided entity so it can be properly
		// handled by the backend
		event.Check.ProxyEntityID = event.Entity.ID
	}

	// From this point we make sure that the agent's entity is used in the event
	// so we provide details like the environment and the organization to the
	// backend
	event.Entity = a.getAgentEntity()
}
