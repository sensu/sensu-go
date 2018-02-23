package types

import (
	"errors"
	"time"
)

// EventFailingState indicates failing check result status
const EventFailingState = "failing"

// EventFlappingState indicates a rapid change in check result status
const EventFlappingState = "flapping"

// EventPassingState indicates successful check result status
const EventPassingState = "passing"

// FixtureEvent returns a testing fixutre for an Event object.
func FixtureEvent(entityID, checkID string) *Event {
	return &Event{
		Timestamp: time.Now().Unix(),
		Entity:    FixtureEntity(entityID),
		Check:     FixtureCheck(checkID),
	}
}

// Validate returns an error if the event does not pass validation tests.
func (e *Event) Validate() error {
	if e.Check == nil || e.Entity == nil {
		return errors.New("malformed event")
	}

	if err := e.Entity.Validate(); err != nil {
		return errors.New("entity " + err.Error())
	}

	if err := e.Check.Validate(); err != nil {
		return errors.New("check " + err.Error())
	}

	for _, hook := range e.Hooks {
		if err := hook.Validate(); err != nil {
			return errors.New("hook " + err.Error())
		}
	}

	return nil
}

// HasCheck determines if an event has check data.
func (e *Event) HasCheck() bool {
	return e.Check != nil
}

// HasMetrics determines if an event has metric data.
func (e *Event) HasMetrics() bool {
	return e.Metrics != nil
}

// IsIncident determines if an event indicates an incident.
func (e *Event) IsIncident() bool {
	return e.HasCheck() && e.Check.Status != 0
}

// IsResolution returns true if an event has just transitionned from an incident
func (e *Event) IsResolution() bool {
	// Try to retrieve the previous status in the check history and verify if it
	// was a non-zero status, therefore indicating a resolution
	isResolution := (len(e.Check.History) > 0 &&
		e.Check.History[len(e.Check.History)-1].Status != 0 &&
		!e.IsIncident())

	return isResolution
}

// IsSilenced determines if an event has any silenced entries
func (e *Event) IsSilenced() bool {
	return len(e.Silenced) > 0
}
