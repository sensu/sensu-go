package v2

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
	assert.Error(t, m.Validate())
	m.Name = "foo"

	// Invalid command
	assert.Error(t, m.Validate())
	m.Command = "echo 'foo'"

	// Invalid namespace
	assert.Error(t, m.Validate())
	m.Namespace = "default"

	// Valid mutator
	assert.NoError(t, m.Validate())
}
