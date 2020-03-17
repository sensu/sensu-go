package v2

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// Invalid namespace
	assert.Error(t, c.Validate())
	c.Namespace = "default"

	// Invalid ttl
	c.Ttl = 10
	assert.Error(t, c.Validate())

	// Invalid output metric format
	c.OutputMetricFormat = "foo"
	assert.Error(t, c.Validate())
	c.OutputMetricFormat = ""

	// Valid check
	c.Ttl = 90
	assert.NoError(t, c.Validate())
}

func TestCheckConfigHasNonNilSubscriptions(t *testing.T) {
	var c CheckConfig
	b, err := json.Marshal(&c)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &c))
	require.NotNil(t, c.Subscriptions)
}

func TestCheckConfigHasNonNilHandlers(t *testing.T) {
	var c CheckConfig
	b, err := json.Marshal(&c)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &c))
	require.NotNil(t, c.Handlers)
}

func TestCheckConfigHasNoEmptyStringsInSub(t *testing.T) {
	c := FixtureCheckConfig("foo")
	c.Subscriptions = append(c.Subscriptions, "demo", "foo")
	assert.NoError(t, c.Validate())
}

func TestCheckConfigErrIfHasEmptyStringsInSub(t *testing.T) {
	c := FixtureCheckConfig("foo")
	c.Subscriptions = append(c.Subscriptions, "")

	assert.Error(t, c.Validate())
}

func TestCheckConfigErrMsgIfHasEmptyStringsInSub(t *testing.T) {
	c := FixtureCheckConfig("foo")
	c.Subscriptions = append(c.Subscriptions, "")

	err := c.Validate()
	assert.EqualError(t, err,  "subscriptions cannot be empty strings")
}

func TestCheckConfigFlapThresholdValidation(t *testing.T) {
	c := FixtureCheckConfig("foo")
	// zero-valued flap threshold is valid
	c.LowFlapThreshold, c.HighFlapThreshold = 0, 0
	assert.NoError(t, c.Validate())

	// low flap threshold < high flap threshold is valid
	c.LowFlapThreshold, c.HighFlapThreshold = 5, 10
	assert.NoError(t, c.Validate())

	// low flap threshold = high flap threshold is invalid
	c.LowFlapThreshold, c.HighFlapThreshold = 10, 10
	assert.Error(t, c.Validate())

	// low flap threshold > high flap threshold is invalid
	c.LowFlapThreshold, c.HighFlapThreshold = 11, 10
	assert.Error(t, c.Validate())
}

func TestSortCheckConfigsByName(t *testing.T) {
	a := FixtureCheckConfig("Abernathy")
	b := FixtureCheckConfig("Bernard")
	c := FixtureCheckConfig("Clementine")
	d := FixtureCheckConfig("Dolores")

	testCases := []struct {
		name     string
		inDir    bool
		inChecks []*CheckConfig
		expected []*CheckConfig
	}{
		{
			name:     "Sorts ascending",
			inDir:    true,
			inChecks: []*CheckConfig{d, c, a, b},
			expected: []*CheckConfig{a, b, c, d},
		},
		{
			name:     "Sorts descending",
			inDir:    false,
			inChecks: []*CheckConfig{d, a, c, b},
			expected: []*CheckConfig{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortCheckConfigsByName(tc.inChecks, tc.inDir))
			assert.EqualValues(t, tc.expected, tc.inChecks)
		})
	}
}
