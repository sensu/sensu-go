package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureRule(t *testing.T) {
	r := FixtureRule("foo")
	assert.Equal(t, "*", r.Type)
	assert.Equal(t, "foo", r.Organization)
	assert.Equal(t, []string{"create", "read", "update", "delete"}, r.Permissions)
}

func TestFixtureRole(t *testing.T) {
	r := FixtureRole("foo", "bar")
	assert.Equal(t, "foo", r.Name)
	assert.NotEmpty(t, r.Rules)
}

func TestRuleValidate(t *testing.T) {
	r := &Rule{Type: "battlestar"}

	// Empty organization
	assert.Error(t, r.Validate())
	r.Organization = "bar"

	// No permissions
	assert.Error(t, r.Validate())
	r.Permissions = []string{"docking"}

	// Invalid permissions
	assert.Error(t, r.Validate())
	r.Permissions = []string{"create"}

	// Valid params
	assert.Equal(t, "battlestar", r.Type)
	assert.Equal(t, "bar", r.Organization)
	assert.Equal(t, []string{"create"}, r.Permissions)
	assert.NoError(t, r.Validate())
}

func TestRoleValidate(t *testing.T) {
	r := FixtureRole("foo", "bar")

	// Valid
	assert.NoError(t, r.Validate())
	assert.Equal(t, "foo", r.Name)

	// Bad name
	r.Name = "FOO/bar/10"
	assert.Error(t, r.Validate())

	// No rules
	r.Rules = []Rule{}
	assert.Error(t, r.Validate())
}
