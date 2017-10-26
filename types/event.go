package types

import "time"

// EventFailingState indicates failing check result status
const EventFailingState = "failing"

// EventFlappingState indicates a rapid change in check result status
const EventFlappingState = "flapping"

// EventPassingState indicates successful check result status
const EventPassingState = "passing"

// An Event is the encapsulating type sent across the Sensu websocket transport.
type Event struct {
	// Timestamp is the time in seconds since the Epoch.
	Timestamp int64 `json:"timestamp"`

	Entity  *Entity  `json:"entity,omitempty"`
	Check   *Check   `json:"check,omitempty"`
	Metrics *Metrics `json:"metrics,omitempty"`
	// Silenced is a list of silenced subscriptions for marking an event as
	// silenced.
	Silenced []string `json:"silenced,omitempty"`
}

// FixtureEvent returns a testing fixutre for an Event object.
func FixtureEvent(entityID, checkID string) *Event {
	return &Event{
		Timestamp: time.Now().Unix(),
		Entity:    FixtureEntity(entityID),
		Check:     FixtureCheck(checkID),
	}
}
