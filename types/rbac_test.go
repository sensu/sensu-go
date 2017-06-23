package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureRule(t *testing.T) {
	r := FixtureRule()
	assert.Equal(t, "*", r.Type)
	assert.Equal(t, []string{"create", "read", "update", "delete"}, r.Permissions)
}

func TestFixtureRole(t *testing.T) {
	r := FixtureRole("foo", "bar")
	assert.Equal(t, "foo", r.Name)
	assert.Equal(t, "bar", r.Organization)
	assert.NotEmpty(t, r.Rules)
}

func TestRuleValidate(t *testing.T) {
	r := &Rule{}

	// Empty organization
	assert.Error(t, r.Validate())

	// No permissions
	r.Type = "battlestar"
	assert.Error(t, r.Validate())

	// Invalid permissions
	r.Permissions = []string{"docking"}
	assert.Error(t, r.Validate())

	r.Permissions = []string{"create"}
	assert.Equal(t, "battlestar", r.Type)
	assert.Equal(t, []string{"create"}, r.Permissions)
	assert.NoError(t, r.Validate())
}

func TestRoleValidate(t *testing.T) {
	r := FixtureRole("foo", "bar")

	// Valid
	assert.NoError(t, r.Validate())
	assert.Equal(t, "foo", r.Name)
	assert.Equal(t, "bar", r.Organization)

	// Bad name
	r.Name = "FOO/bar/10"
	assert.Error(t, r.Validate())

	// Empty org
	r.Organization = "colonial-fleet"
	assert.Error(t, r.Validate())

	// No rules
	r.Rules = []Rule{}
	assert.Error(t, r.Validate())
}
