package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureFilter(t *testing.T) {
	filter := FixtureFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, FilterActionAllow, filter.Action)
	assert.Equal(t, map[string]interface{}{
		"check": map[string]interface{}{
			"team": "ops",
		},
	}, filter.Attributes)
	assert.NoError(t, filter.Validate())
}

func TestFixtureDenyFilter(t *testing.T) {
	filter := FixtureDenyFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, FilterActionDeny, filter.Action)
	assert.Equal(t, map[string]interface{}{
		"check": map[string]interface{}{
			"team": "ops",
		},
	}, filter.Attributes)
	assert.NoError(t, filter.Validate())
}

func TestFilterValidate(t *testing.T) {
	var f Filter

	// Invalid name
	assert.Error(t, f.Validate())
	f.Name = "foo"

	// Invalid action
	assert.Error(t, f.Validate())
	f.Action = "allow"

	// Invalid attributes
	assert.Error(t, f.Validate())
	f.Attributes = map[string]interface{}{
		"check": map[string]interface{}{
			"team": "ops",
		},
	}

	// Invalid organization
	assert.Error(t, f.Validate())
	f.Organization = "default"

	// Invalid environment
	assert.Error(t, f.Validate())
	f.Environment = "default"

	// Valid filter
	assert.NoError(t, f.Validate())
}
