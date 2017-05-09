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
	assert.NotNil(t, h.Validate())
	h.Name = "foo"

	// Invalid type
	assert.NotNil(t, h.Validate())
	h.Type = "pipe"

	// Valid handler
	assert.Nil(t, h.Validate())
}
