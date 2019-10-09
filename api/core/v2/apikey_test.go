package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureAPIKey(t *testing.T) {
	a := FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "bar")
	assert.NoError(t, a.Validate())
	assert.Equal(t, "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", a.Name)
	assert.Equal(t, "bar", a.Username)
	assert.Equal(t, "", a.Namespace)
}

func TestAPIKeyValidate(t *testing.T) {
	a := &APIKey{}

	// Namespace
	a.Namespace = "foo"
	assert.Error(t, a.Validate())
	a.Namespace = ""

	// Empty username
	assert.Error(t, a.Validate())
	a.Username = "bar"

	// Empty name
	assert.Error(t, a.Validate())
	a.Name = "foo"

	// Invalid name
	assert.Error(t, a.Validate())
	a.Name = "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6"

	assert.NoError(t, a.Validate())
	assert.Equal(t, "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", a.Name)
	assert.Equal(t, "bar", a.Username)
	assert.Equal(t, "", a.Namespace)
}
