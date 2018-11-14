package types

import (
	"errors"
	fmt "fmt"
	"net/url"
)

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

	// HandlerGRPCType is a special kind of handler that represents an extension
	HandlerGRPCType = "grpc"
)

// Validate returns an error if the handler does not pass validation tests.
func (h *Handler) Validate() error {
	if err := ValidateName(h.Name); err != nil {
		return errors.New("handler name " + err.Error())
	}

	if err := h.validateType(); err != nil {
		return err
	}

	if h.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

func (h *Handler) validateType() error {
	if h.Type == "" {
		return errors.New("empty handler type")
	}

	switch h.Type {
	case "pipe", "set", "grpc":
		return nil
	case "tcp", "udp":
		return h.Socket.Validate()
	}

	return fmt.Errorf("unknown handler type: %s", h.Type)
}

// Validate returns an error if the handler socket does not pass validation tests.
func (s *HandlerSocket) Validate() error {
	if s == nil {
		return errors.New("tcp and udp handlers need a valid socket")
	}
	if len(s.Host) == 0 {
		return errors.New("socket host undefined")
	}
	if s.Port == 0 {
		return errors.New("socket port undefined")
	}
	return nil
}

// FixtureHandler returns a Handler fixture for testing.
func FixtureHandler(name string) *Handler {
	return &Handler{
		Type:    HandlerPipeType,
		Command: "command",
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
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

// URIPath returns the path component of a Handler URI.
func (h *Handler) URIPath() string {
	return fmt.Sprintf("/handlers/%s", url.PathEscape(h.Name))
}
