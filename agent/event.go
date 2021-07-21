package agent

import (
	"fmt"

	time "github.com/echlebek/timeproxy"
	"github.com/google/uuid"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev1 "github.com/sensu/sensu-go/types/v1"
)

// prepareEvent accepts a partial or complete event and tries to add any missing
// attributes so it can pass validation. An error is returned if it still can't
// pass validation after all these changes
func prepareEvent(a *Agent, event *corev2.Event) error {
	if event == nil {
		return fmt.Errorf("an event must be provided")
	}

	if !event.HasCheck() && !event.HasMetrics() {
		return fmt.Errorf("at least check or metrics need to be instantiated for this event")
	}

	if event.ObjectMeta.Namespace == "" {
		event.ObjectMeta.Namespace = a.config.Namespace
	}

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Make sure the check, if present, has all the required attributes
	if event.HasCheck() {
		if event.Check.Interval == 0 {
			event.Check.Interval = 1
		}

		if event.Check.Namespace == "" {
			event.Check.Namespace = a.config.Namespace
		}

		if event.Check.Executed == 0 {
			event.Check.Executed = time.Now().Unix()
		}

		// The check should pass validation at this point
		if err := event.Check.Validate(); err != nil {
			return err
		}
	}

	// Verify if an entity was provided and that it's not the agent's entity.
	// If so, we have a proxy entity and we need to identify it as the source
	// so it can be properly handled by the backend. Othewise we need to inject
	// the agent's entity into this event
	a.getEntities(event)

	if len(event.ID) == 0 {
		id, err := uuid.NewRandom()
		if err == nil {
			event.ID = id[:]
		}
	}

	// The entity should pass validation at this point
	return event.Entity.Validate()
}

// translateToEvent accepts a 1.x compatible check result
// and attempts to translate it to a 2.x event
func translateToEvent(a *Agent, result corev1.CheckResult, event *corev2.Event) error {
	if result.Name == "" {
		return fmt.Errorf("a check name must be provided")
	}

	if result.Output == "" {
		return fmt.Errorf("a check output must be provided")
	}

	source := result.Source
	if source == "" {
		source = result.Client
	}

	agentEntity := a.getAgentEntity()
	if source == "" || source == agentEntity.Name {
		event.Entity = agentEntity
	} else {
		event.Entity = &corev2.Entity{
			ObjectMeta:  corev2.NewObjectMeta(source, agentEntity.Namespace),
			EntityClass: corev2.EntityProxyClass,
		}
	}

	handlers := result.Handlers
	if len(handlers) == 0 && result.Handler != "" {
		handlers = append(handlers, result.Handler)
	}

	check := &corev2.Check{
		ObjectMeta:    corev2.NewObjectMeta(result.Name, agentEntity.Namespace),
		Status:        result.Status,
		Command:       result.Command,
		Subscriptions: result.Subscribers,
		Interval:      result.Interval,
		Issued:        result.Issued,
		Executed:      result.Executed,
		Duration:      result.Duration,
		Output:        result.Output,
		Handlers:      handlers,
	}

	// add config and check values to the 2.x event
	event.Check = check

	return nil
}
