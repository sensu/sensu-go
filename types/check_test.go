package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckConfigurationValidate(t *testing.T) {
	var c CheckConfiguration

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
	config := c.Configuration

	assert.Equal(t, "check", config.Name)
	assert.NoError(t, config.Validate())

	config.RuntimeAssets = []Asset{
		{Name: "Good", URL: "https://sweet.sweet/good/url.boy"},
	}
	assert.NoError(t, config.Validate())

	config.RuntimeAssets = []Asset{{Name: ""}}
	assert.Error(t, config.Validate())
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
