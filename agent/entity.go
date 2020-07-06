package agent

import (
	time "github.com/echlebek/timeproxy"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/version"
)

func (a *Agent) getAgentEntity() *corev2.Entity {
	a.entityMu.Lock()
	defer a.entityMu.Unlock()
	if a.entityConfig == nil {
		a.entityConfig = a.getLocalEntityConfig()
	}
	state := a.getEntityState()
	entity, err := corev3.V3EntityToV2(a.entityConfig, state)
	if err == nil {
		return entity
	}
	if err != nil {
		state.Metadata = a.entityConfig.Metadata
	}
	entity, err = corev3.V3EntityToV2(a.entityConfig, state)
	if err != nil {
		// unrecoverable error
		panic(err)
	}
	return entity
}

func (a *Agent) getLocalEntityConfig() *corev3.EntityConfig {
	meta := corev2.NewObjectMeta(a.config.AgentName, a.config.Namespace)
	meta.Labels = a.config.Labels
	meta.Annotations = a.config.Annotations
	e := &corev3.EntityConfig{
		EntityClass:       corev2.EntityAgentClass,
		Deregister:        a.config.Deregister,
		Redact:            a.config.Redact,
		Subscriptions:     a.config.Subscriptions,
		User:              a.config.User,
		Metadata:          &meta,
		KeepaliveHandlers: a.config.KeepaliveHandlers,
	}

	if a.config.DeregistrationHandler != "" {
		e.Deregistration = corev2.Deregistration{
			Handler: a.config.DeregistrationHandler,
		}
	}
	return e
}

func (a *Agent) getEntityState() *corev3.EntityState {
	meta := corev2.NewObjectMeta(a.config.AgentName, a.config.Namespace)
	meta.Labels = a.config.Labels
	meta.Annotations = a.config.Annotations
	return &corev3.EntityState{
		Metadata:          &meta,
		SensuAgentVersion: version.Semver(),
		LastSeen:          time.Now().Unix(),
		System:            a.getSystemInfo(),
	}
}

func (a *Agent) getSystemInfo() corev2.System {
	// Retrieve the system info from the agent's cached value
	a.systemInfoMu.RLock()
	defer a.systemInfoMu.RUnlock()
	if a.systemInfo == nil {
		return corev2.System{}
	}
	return *a.systemInfo
}

// getEntities receives an event and verifies if we have a proxy entity, so it
// can be added as the source, and ensures that the event uses the agent's
// entity
func (a *Agent) getEntities(event *corev2.Event) {
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
