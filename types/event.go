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

	return nil
}

// HasMetrics determines if an event has metric data.
func (e *Event) HasMetrics() bool {
	if e.Metrics != nil {
		return true
	}

	return false
}

// IsIncident determines if an event indicates an incident.
func (e *Event) IsIncident() bool {
	if e.Check.Status != 0 {
		return true
	}

	return false
}

// IsResolution returns true if an event has just transitionned from an incident
func (e *Event) IsResolution() bool {
	// Try to retrieve the previous status in the check history and verify if it
	// was a non-zero status, therefore indicating a resolution
	if len(e.Check.History) > 0 && e.Check.History[len(e.Check.History)-1].Status != 0 &&
		!e.IsIncident() {
		return true
	}

	return false
}

// IsSilenced determines if an event has any silenced entries
func (e *Event) IsSilenced() bool {
	if len(e.Silenced) > 0 {
		return true
	}

	return false
}
