package types

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
