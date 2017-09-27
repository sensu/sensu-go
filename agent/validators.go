package agent

import (
	"fmt"
	"time"

	"github.com/sensu/sensu-go/types"
)

// validateEvent validates an event add, if possible, missing information
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
	}

	if event.Check.Config.Interval == 0 {
		event.Check.Config.Interval = 1
	}

	if event.Check.Config.Organization == "" {
		event.Check.Config.Organization = a.config.Organization
	}

	if event.Check.Config.Environment == "" {
		event.Check.Config.Environment = a.config.Environment
	}

	return nil
}
