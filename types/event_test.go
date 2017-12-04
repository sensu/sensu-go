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
}

func TestEventValidate(t *testing.T) {
	event := FixtureEvent("entity", "check")

	event.Check.Config.Name = ""
	assert.Error(t, event.Validate())
	event.Check.Config.Name = "check"

	event.Entity.ID = ""
	assert.Error(t, event.Validate())
	event.Entity.ID = "entity"

	assert.NoError(t, event.Validate())
}
