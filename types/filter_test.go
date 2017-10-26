package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureEventFilter(t *testing.T) {
	filter := FixtureEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionAllow, filter.Action)
	assert.Equal(t, []string{"event.Check.Team == 'ops'"}, filter.Statements)
	assert.NoError(t, filter.Validate())
}

func TestFixtureDenyEventFilter(t *testing.T) {
	filter := FixtureDenyEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionDeny, filter.Action)
	assert.Equal(t, []string{"event.Check.Team == 'ops'"}, filter.Statements)
	assert.NoError(t, filter.Validate())
}

func TestEventFilterValidate(t *testing.T) {
	var f EventFilter

	// Invalid name
	assert.Error(t, f.Validate())
	f.Name = "foo"

	// Invalid action
	assert.Error(t, f.Validate())
	f.Action = "allow"

	// Invalid attributes
	assert.Error(t, f.Validate())
	f.Statements = []string{"event.Check.Team == 'ops'"}

	// Invalid organization
	assert.Error(t, f.Validate())
	f.Organization = "default"

	// Invalid environment
	assert.Error(t, f.Validate())
	f.Environment = "default"

	// Valid filter
	assert.NoError(t, f.Validate())
}

func TestValidateStatements(t *testing.T) {
	// Valid statement
	statements := []string{"10 > 0"}
	assert.NoError(t, validateStatements(statements))

	// Invalid statement
	statements = []string{"10. 0"}
	assert.Error(t, validateStatements(statements))

	// Forbidden modifier token
	statements = []string{"10 + 2 > 0"}
	assert.Error(t, validateStatements(statements))

}
