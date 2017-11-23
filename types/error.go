package types

// GetOrg refers to the organization the check belongs to
func (e *Error) GetOrg() string {
	return e.Organization
}

// GetEnv refers to the organization the check belongs to
func (e *Error) GetEnv() string {
	return e.Environment
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
