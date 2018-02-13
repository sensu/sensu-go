package agent

import (
	"fmt"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/v1"
)

// prepareEvent accepts a partial or complete event and tries to add any missing
// attributes so it can pass validation. An error is returned if it still can't
// pass validation after all these changes
func prepareEvent(a *Agent, event *types.Event) error {
	if event == nil {
		return fmt.Errorf("an event must be provided")
	}

	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Make sure the check has all required attributes
	if event.Check.Interval == 0 {
		event.Check.Interval = 1
	}

	if event.Check.Organization == "" {
		event.Check.Organization = a.config.Organization
	}

	if event.Check.Environment == "" {
		event.Check.Environment = a.config.Environment
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

// translateToEvent accepts a 1.x compatible check result
// and attempts to translate it to a 2.x event
func translateToEvent(a *Agent, result v1.CheckResult, event *types.Event) error {
	if result.Name == "" {
		return fmt.Errorf("a check name must be provided")
	}

	if result.Output == "" {
		return fmt.Errorf("a check output must be provided")
	}

	agentEntity := a.getAgentEntity()
	if result.Client == "" || result.Client == agentEntity.ID {
		event.Entity = agentEntity
	} else {
		event.Entity = &types.Entity{
			ID:    result.Client,
			Class: types.EntityProxyClass,
		}
	}

	check := &types.Check{
		Status:        result.Status,
		Command:       result.Command,
		Subscriptions: result.Subscribers,
		Interval:      result.Interval,
		Name:          result.Name,
		Issued:        result.Issued,
		Executed:      result.Executed,
		Duration:      result.Duration,
		Output:        result.Output,
	}
	check.SetExtendedAttributes(result.GetExtendedAttributes())

	// add config and check values to the 2.x event
	event.Check = check

	return nil
}
