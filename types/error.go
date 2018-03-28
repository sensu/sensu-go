package types

import (
	"time"
)

// GetOrganization returns the organization the entity is associated with.
func (perr *Error) GetOrganization() string {
	return perr.Event.Entity.GetOrganization()
}

// GetEnvironment returns the environmen the entity is associated with.
func (perr *Error) GetEnvironment() string {
	return perr.Event.Entity.GetEnvironment()
}

// FixtureError returns a testing fixture for an Error object.
func FixtureError(name string, message string) *Error {
	event := FixtureEvent("agent", "check")

	return &Error{
		Name:      name,
		Component: "pipelined",
		Message:   message,
		Event:     *event,
		Timestamp: time.Now().Unix(),
	}
}
