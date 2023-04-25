package agent

import (
	"fmt"

	time "github.com/echlebek/timeproxy"
	"github.com/google/uuid"

	corev2 "github.com/sensu/core/v2"
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
		event.Check.ProcessedBy = a.config.AgentName
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
