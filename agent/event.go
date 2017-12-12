package agent

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
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

// translateToEvent accepts a 1.x compatible payload
// and attempts to translate it to a 2.x event
func translateToEvent(payload map[string]interface{}, event *types.Event) error {
	var checkConfig types.CheckConfig
	var check types.Check

	if payload == nil {
		return fmt.Errorf("a payload must be provided")
	}
	// dump relevant payload values into 2.x config struct fields
	if err := mapstructure.Decode(payload, &checkConfig); err != nil {
		return fmt.Errorf("error translating check config")
	}
	// dump relevant payload values into 2.x check struct fields
	if err := mapstructure.Decode(payload, &check); err != nil {
		return fmt.Errorf("error translating check")
	}

	// add config and check values to the 2.x event
	check.Config = &checkConfig
	event.Check = &check

	configVal := reflect.Indirect(reflect.ValueOf(check.Config))
	checkVal := reflect.Indirect(reflect.ValueOf(event.Check))
	for mapKey := range payload {
		_, existsInConfig := configVal.Type().FieldByName(strings.Title(mapKey))
		_, existsInCheck := checkVal.Type().FieldByName(strings.Title(mapKey))
		// if the field does not exist in 2.x check or 2.x config structs,
		// add the custom attribute to the config only
		if !(existsInConfig || existsInCheck) {
			if err := dynamic.SetField(check.Config, mapKey, payload[mapKey]); err != nil {
				return fmt.Errorf("error translating custom attributes")
			}
		}
	}

	return nil
}
