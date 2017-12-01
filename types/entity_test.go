package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityValidate(t *testing.T) {
	var e Entity

	// Invalid ID
	assert.Error(t, e.Validate())
	e.ID = "foo"

	// Invalid class
	assert.Error(t, e.Validate())
	e.Class = "agent"

	// Invalid organization
	assert.Error(t, e.Validate())
	e.Organization = "default"

	// Invalid environment
	assert.Error(t, e.Validate())
	e.Environment = "default"

	// Valid entity
	assert.NoError(t, e.Validate())
}

func TestFixtureEntityIsValid(t *testing.T) {
	e := FixtureEntity("entity")
	assert.Equal(t, "entity", e.ID)
	assert.NoError(t, e.Validate())
}

func TestEntityGet(t *testing.T) {
	e := FixtureEntity("entity")
	e.ID = "Test"
	e.ExtendedAttributes = []byte(`{}`)

	// Simple value
	val, err := e.Get("ID")
	require.NoError(t, err)
	assert.EqualValues(t, "Test", val)
}
