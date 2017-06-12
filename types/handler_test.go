package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureHandler(t *testing.T) {
	handler := FixtureHandler("handler")
	assert.Equal(t, "handler", handler.Name)
	assert.NoError(t, handler.Validate())
}

func TestFixtureSetHandler(t *testing.T) {
	handler := FixtureSetHandler("handler")
	assert.Equal(t, "handler", handler.Name)
	assert.NoError(t, handler.Validate())
}

func TestFixtureSocketHandler(t *testing.T) {
	handler := FixtureSocketHandler("handler", "tcp")
	assert.Equal(t, "handler", handler.Name)
	assert.Equal(t, "tcp", handler.Type)
	assert.NotNil(t, handler.Socket.Host)
	assert.NotNil(t, handler.Socket.Port)
	assert.NoError(t, handler.Validate())
}

func TestHandlerValidate(t *testing.T) {
	var h Handler

	// Invalid name
	assert.Error(t, h.Validate())
	h.Name = "foo"

	// Invalid type
	assert.Error(t, h.Validate())
	h.Type = "pipe"

	// Invalid organization
	assert.Error(t, h.Validate())
	h.Organization = "default"

	// Valid handler
	assert.NoError(t, h.Validate())
}
