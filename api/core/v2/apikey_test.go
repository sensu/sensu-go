package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixutreAPIKey(t *testing.T) {
	a := FixtureAPIKey("foo", "bar")
	assert.NoError(t, a.Validate())
	assert.Equal(t, "foo", a.Name)
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

	assert.NoError(t, a.Validate())
	assert.Equal(t, "foo", a.Name)
	assert.Equal(t, "bar", a.Username)
	assert.Equal(t, "", a.Namespace)
}
