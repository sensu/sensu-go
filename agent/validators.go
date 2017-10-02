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

	// Use the agent entity if none was passed
	if event.Entity == nil {
		event.Entity = a.getAgentEntity()
	} else {
		// Make sure the entity has all required attributes
		if event.Entity.Class == "" {
			event.Entity.Class = "proxy"
		}

		if event.Entity.Organization == "" {
			event.Entity.Organization = a.config.Organization
		}

		if event.Entity.Environment == "" {
			event.Entity.Environment = a.config.Environment
		}
	}

	// The entity should pass validation at this point
	if err := event.Entity.Validate(); err != nil {
		return err
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
	return event.Check.Validate()
}
