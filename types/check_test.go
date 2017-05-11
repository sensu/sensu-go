package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValidate(t *testing.T) {
	var c Check

	// Invalid name
	assert.Error(t, c.Validate())
	c.Name = "foo"

	// Invalid interval
	assert.Error(t, c.Validate())
	c.Interval = 60

	// Invalid command
	assert.Error(t, c.Validate())
	c.Command = "echo 'foo'"

	// Valid check
	assert.NoError(t, c.Validate())
}

func TestFixtureCheckIsValid(t *testing.T) {
	c := FixtureCheck("check")
	assert.Equal(t, "check", c.Name)
	assert.NoError(t, c.Validate())
}

func TestMergeWith(t *testing.T) {
	originalCheck := FixtureCheck("check")
	originalCheck.Status = 1

	newCheck := FixtureCheck("check")
	newCheck.History = []CheckHistory{}

	newCheck.MergeWith(originalCheck)

	assert.NotEmpty(t, newCheck.History)
	// History has a length of 21, so just index it directly. jfc.
	assert.Equal(t, 1, newCheck.History[20].Status)
}
