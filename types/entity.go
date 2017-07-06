package types

import "errors"

// Entity is the Entity supplying the event. The default Entity for any
// Event is the running Agent process--if the Event is sent by an Agent.
type Entity struct {
	ID               string         `json:"id"`
	Class            string         `json:"class"`
	System           System         `json:"system,omitempty"`
	Subscriptions    []string       `json:"subscriptions,omitempty"`
	LastSeen         int64          `json:"last_seen,omitempty"`
	Deregister       bool           `json:"deregister"`
	Deregistration   Deregistration `json:"deregistration"`
	KeepaliveTimeout uint           `json:"keepalive_timeout"`
	Organization     string         `json:"organization"`
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

// Deregistration contains configuration for Sensu entity de-registration.
type Deregistration struct {
	Handler string `json:"handler"`
}

// Validate returns an error if the entity is invalid.
func (e *Entity) Validate() error {
	if err := ValidateName(e.ID); err != nil {
		return errors.New("entity id " + err.Error())
	}

	if err := ValidateName(e.Class); err != nil {
		return errors.New("entity class " + err.Error())
	}

	if e.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureEntity(id string) *Entity {
	return &Entity{
		ID:               id,
		Class:            "host",
		Subscriptions:    []string{"subscription"},
		Organization:     "default",
		KeepaliveTimeout: 120,
	}
}
