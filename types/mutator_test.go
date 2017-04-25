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
