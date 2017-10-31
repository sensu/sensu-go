package types

import "time"

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
