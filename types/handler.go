package types

import "errors"

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

// GetOrg refers to the organization the handler belongs to
func (h *Handler) GetOrg() string {
	return h.Organization
}

// GetEnv refers to the organization the handler belongs to
func (h *Handler) GetEnv() string {
	return h.Environment
}

// FixtureHandler returns a Handler fixture for testing.
func FixtureHandler(name string) *Handler {
	return &Handler{
		Name:         name,
		Type:         "pipe",
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
