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

func TestHandlerValidate(t *testing.T) {
	var h Handler

	// Invalid name
	assert.Error(t, h.Validate())
	h.Name = "foo"

	// Invalid type
	assert.Error(t, h.Validate())
	h.Type = "pipe"

	// Valid handler
	assert.NoError(t, h.Validate())
}
