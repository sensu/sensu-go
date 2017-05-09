package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityValidate(t *testing.T) {
	var e Entity

	// Invalid ID
	assert.NotNil(t, e.Validate())
	e.ID = "foo"

	// Invalid class
	assert.NotNil(t, e.Validate())
	e.Class = "agent"

	// Valid entity
	assert.Nil(t, e.Validate())
}

func TestFixtureEntityIsValid(t *testing.T) {
	e := FixtureEntity("entity")
	assert.Equal(t, "entity", e.ID)
	assert.NoError(t, e.Validate())
}
