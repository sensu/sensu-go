package logging

import (
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// EventFields populates a map with fields containing relevant information about
// an event for logging
func EventFields(event *types.Event) map[string]interface{} {
	// Ensure the entity is present
	if event.Entity == nil {
		return map[string]interface{}{}
	}

	fields := logrus.Fields{
		"entity":       event.Entity.ID,
		"environment":  event.Entity.Environment,
		"organization": event.Entity.Organization,
	}

	// The event might not have a check if it's a metric so verify that before
	// adding the check name
	if event.HasCheck() {
		fields["check"] = event.Check.Name
	}

	return fields
}
