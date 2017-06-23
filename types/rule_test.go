package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureRule(t *testing.T) {
	r := FixtureRule("foo")
	assert.Equal(t, "foo", r.Organization)
	assert.Equal(t, "*", r.Type)
	assert.Equal(t, []string{"create", "read", "update", "delete"}, r.Permissions)
}

func TestRuleValidate(t *testing.T) {
	r := &Rule{}

	// Empty organization
	assert.Error(t, r.Validate())

	// Empty type
	r.Organization = "colonial fleet"
	assert.Error(t, r.Validate())

	// No permissions
	r.Type = "battlestar"
	assert.Error(t, r.Validate())

	// Invalid permissions
	r.Permissions = []string{"docking"}
	assert.Error(t, r.Validate())

	r.Permissions = []string{"create"}
	assert.Equal(t, "colonial fleet", r.Organization)
	assert.Equal(t, "battlestar", r.Type)
	assert.Equal(t, []string{"create"}, r.Permissions)
	assert.NoError(t, r.Validate())
}
