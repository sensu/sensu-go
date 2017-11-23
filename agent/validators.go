package agent

import (
	"fmt"
	"time"

	"github.com/sensu/sensu-go/types"
)

// validateEvent validates an event and tries to add missing attributes
func validateEvent(a *Agent, event *types.Event) error {
	if event == nil {
		return fmt.Errorf("an event must be provided")
	}

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Make sure the check has all required attributes
	if event.Check.Config.Interval == 0 {
		event.Check.Config.Interval = 1
	}

	if event.Check.Config.Organization == "" {
		event.Check.Config.Organization = a.config.Organization
	}

	if event.Check.Config.Environment == "" {
		event.Check.Config.Environment = a.config.Environment
	}

	if event.Check.Executed == 0 {
		event.Check.Executed = time.Now().Unix()
	}

	// The check should pass validation at this point
	if err := event.Check.Validate(); err != nil {
		return err
	}

	// Verify if an entity was provided and that it's not the agent's entity.
	// If so, we have a proxy entity and we need to identify it as the source
	// so it can be properly handled by the backend. Othewise we need to inject
	// the agent's entity into this event
	a.getEntities(event)

	// The entity should pass validation at this point
	return event.Entity.Validate()
}
