package types

import "errors"

// Entity is the Entity supplying the event. The default Entity for any
// Event is the running Agent process--if the Event is sent by an Agent.
type Entity struct {
	ID            string   `json:"id"`
	Class         string   `json:"class"`
	System        System   `json:"system,omitempty"`
	Subscriptions []string `json:"subscriptions,omitempty"`
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
	MAC       string   `json:"mac,omitempty"`
	Addresses []string `json:"addresses"`
}

// Validate returns an error if the entity is invalid.
func (e *Entity) Validate() error {
	if e.ID == "" {
		return errors.New("entity id must not be empty")
	}

	if e.Class == "" {
		return errors.New("entity string must not be empty")
	}

	return nil
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureEntity(id string) *Entity {
	return &Entity{
		ID:            id,
		Class:         "host",
		Subscriptions: []string{"subscription"},
	}
}
