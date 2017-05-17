package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureUser(t *testing.T) {
	u := FixtureUser("foo")
	assert.Equal(t, "foo", u.Username)
}
