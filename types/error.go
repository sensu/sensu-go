package types

// GetOrganization returns the organization the entity is associated with.
func (perr *types.Error) GetOrganization() string {
	return perr.Entity.Organization
}

// GetEnvironment returns the environmen the entity is associated with.
func (perr *types.Error) GetEnvironment() string {
	return perr.Entity.Environment
}

// FixtureError returns a testing fixture for an Error object.
func FixtureError(name string, message string) *Error {
	event := FixtureEvent("agent", "check")

	return &Error{
		Name:         name,
		Organization: "default",
		Environment:  "default",
		Component:    "pipelined",
		Message:      message,
		Event:        *event,
	}
}
