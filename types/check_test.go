package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValidate(t *testing.T) {
	var c Check

	// Invalid status
	c.Status = -1
	assert.Error(t, c.Validate())
	c.Status = 0

	// Valid w/o config
	assert.NoError(t, c.Validate())
	c.Config = &CheckConfig{
		Name: "test",
	}

	// Invalid w/ bad config
	assert.Error(t, c.Validate())
	c.Config = &CheckConfig{
		Name:         "test",
		Interval:     10,
		Command:      "yes",
		Organization: "default",
	}

	// Valid check
	assert.NoError(t, c.Validate())
}

func TestCheckConfig(t *testing.T) {
	var c CheckConfig

	// Invalid name
	assert.Error(t, c.Validate())
	c.Name = "foo"

	// Invalid interval
	assert.Error(t, c.Validate())
	c.Interval = 60

	// Invalid command
	assert.Error(t, c.Validate())
	c.Command = "echo 'foo'"

	// Invalid organization
	assert.Error(t, c.Validate())
	c.Organization = "default"

	// Valid check
	assert.NoError(t, c.Validate())
}

func TestFixtureCheckIsValid(t *testing.T) {
	c := FixtureCheck("check")
	config := c.Config

	assert.Equal(t, "check", config.Name)
	assert.NoError(t, config.Validate())

	config.RuntimeAssets = []Asset{
		*FixtureAsset("good"),
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
