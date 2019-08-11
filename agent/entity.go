package agent

import (
	time "github.com/echlebek/timeproxy"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

func (a *Agent) getAgentEntity() *types.Entity {
	if a.entity == nil {
		meta := v2.NewObjectMeta(a.config.AgentName, a.config.Namespace)
		meta.Labels = a.config.Labels
		meta.Annotations = a.config.Annotations
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
		// Only do this if the entity name isn't empty, as we use empty entity names
		// to signify to the backend that the check was submitted via the agent API,
		// not the result of a previously scheduled check. This is important as it
		// allows the backend to differentiate between events received for
		// prevously scheduled checks for since-deleted entities, and events submitted
		// via the agent API which should create a new entity if one doesn't exist.
		if event.Entity.Name != "" {
			event.Check.ProxyEntityName = event.Entity.Name
		}
	}

	// From this point we make sure that the agent's entity data is used in the event
	// so we provide details like the namespace to the backend
	agentEntity := a.getAgentEntity()

	// If we are dealing with a proxy entity, don't add the Agent's name to it.
	// Having the agent's entity embedded within the event prevents the detection
	// (and subsequent dropping) of events submitted for deleted entities, as well
	// as properly handling malformed events where the proxy_entity_name field
	// doesn't match the entity name embedded in the event.
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		// This may apply in different ways which require slightly different handling:
		//   1. Proxy-style entity is provided in the JSON payload of the API
		//      (ie: { "entity_class": "proxy", "metadata": { "namespace": "default", "name": "foo" }}
		//   2. No entity is provided in the payload
		//
		// Case 1
		if event.Entity == nil {
			// Copy the necessary entity attributes to pass validation
			event.Entity = &types.Entity{
				ObjectMeta: types.ObjectMeta{
					Namespace: agentEntity.Namespace,
				},
			}
			// Case 2
		} else {
			// Namespace is needed by the backend, everything else in the event is looked
			// up from the store, or if the entity is being created, set from defaults,
			// so we don't need to modify anything else.
			if event.Entity.Namespace == "" {
				event.Entity.Namespace = agentEntity.Namespace
			}
		}
	} else {
		// Send as the agent's entity directly
		event.Entity = agentEntity
	}
}
