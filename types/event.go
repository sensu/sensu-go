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
