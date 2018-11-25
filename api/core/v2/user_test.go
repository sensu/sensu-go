package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureUser(t *testing.T) {
	u := FixtureUser("foo")
	assert.NoError(t, u.Validate())
	assert.Equal(t, "foo", u.Username)
	assert.Contains(t, u.Groups, "default")
}

func TestUserValidate(t *testing.T) {
	u := &User{}

	// Empty username
	assert.Error(t, u.Validate())

	u = FixtureUser("foo")
	assert.Equal(t, "foo", u.Username)
	assert.NoError(t, u.Validate())
}

func TestUserValidatePassword(t *testing.T) {
	u := &User{}

	// Empty password
	assert.Error(t, u.ValidatePassword())

	// Too short password
	u = FixtureUser("foo")
	u.Password = "123"
	assert.Error(t, u.ValidatePassword())

	u.Password = "P@ssw0rd!"
	assert.NoError(t, u.ValidatePassword())
}
