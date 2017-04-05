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

	// Pipe contains configuration for a pipe handler.
	Pipe HandlerPipe `json:"pipe,omitempty"`

	// Handlers is a list of handlers for a handler set.
	Handlers []string `json:"handlers,omitempty"`
}

// HandlerPipe contains configuration for a pipe handler.
type HandlerPipe struct {
	// Command is the command to be executed.
	Command string `json:"command"`

	// Timeout is the command execution timeout in seconds.
	Timeout int `json:"timeout"`
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
