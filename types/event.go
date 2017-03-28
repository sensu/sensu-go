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

	Entity *Entity `json:"entity,omitempty"`
	Check  *Check  `json:"check,omitempty"`
}

// Entity is the Entity supplying the event. The default Entity for any
// Event is the running Agent process--if the Event is sent by an Agent.
type Entity struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	System System `json:"system,omitempty"`
}

// System contains information about the system that the Agent process
// is running on, used for additional Entity context.
type System struct {
	Hostname        string  `json:"hostname"`
	OS              string  `json:"os"`
	Platform        string  `json:"platform"`
	PlatformFamily  string  `json:"platform_family"`
	PlatformVersion string  `json:"platform_version"`
	Network         Network `json:"network"`
}

// Network contains information about the system network interfaces
// that the Agent process is running on, used for additional Entity
// context.
type Network struct {
	Interfaces []NetworkInterface `json:"interfaces"`
}

// NetworkInterface contains information about a system network
// interface.
type NetworkInterface struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac"`
	Addresses []string `json:"addresses"`
}
