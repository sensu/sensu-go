package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureSilenced(t *testing.T) {
	s := FixtureSilenced("test_subscription:test_check")
	s.Expire = 60
	s.ExpireOnResolve = true
	s.Creator = "creator@example.com"
	s.Reason = "test reason"
	s.Organization = "default"
	s.Environment = "default"
	assert.NotNil(t, s)
	assert.NotNil(t, s.ID)
	assert.Equal(t, "test_subscription:test_check", s.ID)
	assert.NotNil(t, s.Expire)
	assert.NotNil(t, s.ExpireOnResolve)
	assert.NotNil(t, s.Expire)
	assert.NotNil(t, s.Creator)
	assert.NotNil(t, s.Check)
	assert.NotNil(t, s.Reason)
	assert.NotNil(t, s.Subscription)
	assert.NotNil(t, s.Organization)
	assert.NotNil(t, s.Environment)

	s = FixtureSilenced("entity:test_subscription:test_check")
	assert.Equal(t, "entity:test_subscription", s.Subscription)
	assert.Equal(t, "test_check", s.Check)
}

// Validation should fail when we don't provide a CheckName or Subscription
func TestSilencedValidate(t *testing.T) {
	var s Silenced
	assert.Error(t, s.Validate())
}
