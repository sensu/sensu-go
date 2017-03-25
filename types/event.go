package types

// KeepaliveType is the message type sent for keepalives--which are just an
// event without a Check or Metrics section.
const KeepaliveType = "keepalive"

// EventType is the message type string for events.
const EventType = "event"

// An Event is the encapsulating type sent across the Sensu websocket transport.
type Event struct {
	// Timestamp is the time in seconds since the Epoch.
	Timestamp int64 `json:"timestamp"`

	// Entity is the Entity supplying the event. The default Entity for any
	// Event is the running Agent process--if the Event is sent by an Agent.
	Entity *Entity `json:"entity,omitempty"`
}

// Entity provides Event context.
type Entity struct {
	ID string
	System
}

// System contains information about the system that the Agent process
// is running on, used for additional Entity context.
type System struct {
	Hostname string
}
