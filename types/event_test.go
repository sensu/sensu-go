package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureEventIsValid(t *testing.T) {
	e := FixtureEvent("entity", "check")
	assert.NotNil(t, e)
	assert.NotNil(t, e.Entity)
	assert.NotNil(t, e.Check)
	assert.False(t, e.Silenced)
}
