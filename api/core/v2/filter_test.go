package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureEventFilter(t *testing.T) {
	filter := FixtureEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionAllow, filter.Action)
	assert.Equal(t, []string{"event.check.team == 'ops'"}, filter.Expressions)
	assert.NoError(t, filter.Validate())
}

func TestFixtureDenyEventFilter(t *testing.T) {
	filter := FixtureDenyEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionDeny, filter.Action)
	assert.Equal(t, []string{"event.check.team == 'ops'"}, filter.Expressions)
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
	f.Expressions = []string{"event.check.team == 'ops'"}

	// Invalid namespace
	assert.Error(t, f.Validate())
	f.Namespace = "default"

	// Valid filter
	assert.NoError(t, f.Validate())
}
