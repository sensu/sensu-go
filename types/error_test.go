package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureError(t *testing.T) {
	err := FixtureError("handler", "handler failed to execute")
	assert.Equal(t, "handler", err.Name)
	assert.Contains(t, err.Message, "handler failed")
}
