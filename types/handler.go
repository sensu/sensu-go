package types

import "errors"

// A Handler is a handler specification.
type Handler struct {
	// Name is the unique identifier for a handler.
	Name string `json:"name"`

	// Type is the handler type, i.e. pipe.
	Type string `json:"type"`

	// Mutator is the handler event data mutator.
	Mutator string `json:"mutator,omitempty"`

	// Command is the command to be executed for a pipe handler.
	Command string `json:"command,omitempty"`

	// Timeout is the handler timeout in seconds.
	Timeout int `json:"timeout"`

	Socket HandlerSocket `json:"socket,omitempty"`

	// Handlers is a list of handlers for a handler set.
	Handlers []string `json:"handlers,omitempty"`
}

// HandlerSocket contains configuration for a TCP or UDP handler.
type HandlerSocket struct {
	// Host is the socket peer address.
	Host string `json:"host"`

	// Port is the socket peer port.
	Port int `json:"port"`
}

// Validate returns an error if the handler does not pass validation tests.
func (h *Handler) Validate() error {
	if h.Type == "" {
		return errors.New("must have a type")
	}

	if h.Name == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}
