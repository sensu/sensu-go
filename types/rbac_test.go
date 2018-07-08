package types

import (
	"types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureRule(t *testing.T) {
	r := FixtureRule("acme", "dev")
	assert.Equal(t, "*", r.Type)
	assert.Equal(t, "acme", r.Organization)
	assert.Equal(t, "dev", r.Environment)
	assert.Equal(t, []string{"create", "read", "update", "delete"}, r.Permissions)
}

func TestFixtureRole(t *testing.T) {
	r := FixtureRole("foo", "acme", "dev")
	assert.Equal(t, "foo", r.Name)
	assert.NotEmpty(t, r.Rules)
}

func TestRuleValidate(t *testing.T) {
	r := &Rule{Type: "battlestar"}

	// Empty environment
	assert.Error(t, r.Validate())
	r.Environment = "dev"

	// Empty organization
	assert.Error(t, r.Validate())
	r.Organization = "acme"

	// No permissions
	assert.Error(t, r.Validate())
	r.Permissions = []string{"docking"}

	// Invalid permissions
	assert.Error(t, r.Validate())
	r.Permissions = []string{"create"}

	// Valid params
	assert.Equal(t, "battlestar", r.Type)
	assert.Equal(t, "dev", r.Environment)
	assert.Equal(t, "acme", r.Organization)
	assert.Equal(t, []string{"create"}, r.Permissions)
	assert.NoError(t, r.Validate())

	// Wildcard org
	r.Organization = types.OrganizationTypeAll
	assert.NoError(t, r.Validate())
}

func TestRoleValidate(t *testing.T) {
	r := FixtureRole("foo", "acme", "dev")

	// Valid
	assert.NoError(t, r.Validate())
	assert.Equal(t, "foo", r.Name)

	// Bad name
	r.Name = "FOO/bar/10"
	assert.Error(t, r.Validate())

	// Bad rules
	r.Rules = []Rule{{Type: "sdfadfsadsfasdf@##@$!@$"}}
	assert.Error(t, r.Validate())
}
