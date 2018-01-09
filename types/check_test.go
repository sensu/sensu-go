package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Environment:  "default",
		Organization: "default",
		Ttl:          30,
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

	// Invalid cron
	assert.Error(t, c.Validate())
	c.Cron = "0 30 * * * *"

	// Invalid command
	assert.Error(t, c.Validate())
	c.Command = "echo 'foo'"

	// Invalid organization
	assert.Error(t, c.Validate())
	c.Organization = "default"

	// Invalid environment
	assert.Error(t, c.Validate())
	c.Environment = "default"

	// Invalid ttl
	c.Ttl = 10
	assert.Error(t, c.Validate())

	// Valid check
	c.Ttl = 90
	assert.NoError(t, c.Validate())
}

func TestFixtureCheckIsValid(t *testing.T) {
	c := FixtureCheck("check")
	config := c.Config

	assert.Equal(t, "check", config.Name)
	assert.NoError(t, config.Validate())

	config.RuntimeAssets = []string{"good"}
	assert.NoError(t, config.Validate())

	config.RuntimeAssets = []string{"BAD--a!!!---ASDFASDF$$$$"}
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
	assert.Equal(t, int32(1), newCheck.History[20].Status)
}

func TestExtendedAttributes(t *testing.T) {
	type getter interface {
		Get(string) (interface{}, error)
	}
	check := FixtureCheck("chekov")
	check.Config.SetExtendedAttributes([]byte(`{"foo":{"bar":42,"baz":9001}}`))
	g, err := check.Config.Get("foo")
	require.NoError(t, err)
	v, err := g.(getter).Get("bar")
	require.NoError(t, err)
	require.Equal(t, 42.0, v)
}
