package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureMutator(t *testing.T) {
	fixture := FixtureMutator("fixture")
	assert.Equal(t, "fixture", fixture.Name)
	assert.NoError(t, fixture.Validate())
}

func TestMutatorValidate(t *testing.T) {
	var m Mutator

	// Invalid name
	assert.NotNil(t, m.Validate())
	m.Name = "foo"

	// Invalid command
	assert.NotNil(t, m.Validate())
	m.Command = "echo 'foo'"

	// Valid mutator
	assert.Nil(t, m.Validate())
}
