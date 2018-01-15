package types

import "errors"

const (
	// HandlerPipeType represents handlers that pipes event data // into arbitrary
	// commands via STDIN
	HandlerPipeType = "pipe"

	// HandlerSetType represents handlers that groups event handlers, making it
	// easy to manage groups of actions that should be executed for certain types
	// of events.
	HandlerSetType = "set"

	// HandlerTCPType represents handlers that send event data to a remote TCP
	// socket
	HandlerTCPType = "tcp"

	// HandlerUDPType represents handlers that send event data to a remote UDP
	// socket
	HandlerUDPType = "udp"
)

// Validate returns an error if the handler does not pass validation tests.
func (h *Handler) Validate() error {
	if err := ValidateName(h.Name); err != nil {
		return errors.New("handler name " + err.Error())
	}

	if err := validateHandlerType(h.Type); err != nil {
		return errors.New("handler type " + err.Error())
	}

	if h.Environment == "" {
		return errors.New("environment must be set")
	}

	if h.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// FixtureHandler returns a Handler fixture for testing.
func FixtureHandler(name string) *Handler {
	return &Handler{
		Name:         name,
		Type:         HandlerPipeType,
		Command:      "command",
		Environment:  "default",
		Organization: "default",
	}
}

// FixtureSocketHandler returns a Handler fixture for testing.
func FixtureSocketHandler(name string, proto string) *Handler {
	handler := FixtureHandler(name)
	handler.Type = proto
	handler.Socket = &HandlerSocket{
		Host: "127.0.0.1",
		Port: 3001,
	}
	return handler
}

// FixtureSetHandler returns a Handler fixture for testing.
func FixtureSetHandler(name string, handlers ...string) *Handler {
	handler := FixtureHandler(name)
	handler.Handlers = handlers
	return handler
}
