package v2

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckValidate(t *testing.T) {
	var c Check

	// Invalid interval
	c.Interval = 0
	assert.Error(t, c.Validate())
	c.Interval = 10

	c.Name = "test"

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

func TestScheduleValidation(t *testing.T) {
	c := FixtureCheck("check")

	// Fixture comes with valid interval-based schedule
	assert.NoError(t, c.Validate())

	c.Cron = "* * * * *"
	assert.Error(t, c.Validate())

	c.Interval = 0
	assert.NoError(t, c.Validate())

	c.Cron = "this is an invalid cron"
	assert.Error(t, c.Validate())
}

func TestFixtureCheckIsValid(t *testing.T) {
	c := FixtureCheck("check")

	assert.Equal(t, "check", c.Name)
	assert.NoError(t, c.Validate())

	c.RuntimeAssets = []string{"good"}
	assert.NoError(t, c.Validate())

	c.RuntimeAssets = []string{"BAD--a!!!---ASDFASDF$$$$"}
	assert.Error(t, c.Validate())
}

func TestMergeWith(t *testing.T) {
	originalCheck := FixtureCheck("check")
	originalCheck.Status = 1

	newCheck := FixtureCheck("check")
	newCheck.History = []CheckHistory{}

	newCheck.MergeWith(originalCheck)

	assert.NotEmpty(t, newCheck.History)
	assert.Equal(t, newCheck.Status, newCheck.History[20].Status)
}

func TestProxyRequestsValidate(t *testing.T) {
	var p ProxyRequests

	// Invalid splay coverage
	p.SplayCoverage = 150
	assert.Error(t, p.Validate())
	p.SplayCoverage = 0

	// Invalid splay and splay coverage
	p.Splay = true
	assert.Error(t, p.Validate())
	p.SplayCoverage = 90

	p.EntityAttributes = []string{`entity.EntityClass == "proxy"`}

	// Valid proxy request
	assert.NoError(t, p.Validate())
}

func TestOutputMetricFormatValidate(t *testing.T) {
	assert.NoError(t, ValidateOutputMetricFormat("nagios_perfdata"))
	assert.NoError(t, ValidateOutputMetricFormat(NagiosOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(GraphiteOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(InfluxDBOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(OpenTSDBOutputMetricFormat))
	assert.Error(t, ValidateOutputMetricFormat("anything_else"))
	assert.Error(t, ValidateOutputMetricFormat("NAGIOS_PERFDATA"))
}

func TestFixtureProxyRequests(t *testing.T) {
	p := FixtureProxyRequests(true)

	assert.Equal(t, true, p.Splay)
	assert.Equal(t, uint32(90), p.SplayCoverage)
	assert.NoError(t, p.Validate())

	p.SplayCoverage = 0
	assert.Error(t, p.Validate())
}

func TestCheckHasZeroIssuedMarshaled(t *testing.T) {
	check := FixtureCheck("foo")
	check.Issued = 0
	b, err := json.Marshal(check)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["issued"]; !ok {
		t.Error("issued not present")
	}
}

func TestCheckHasNonNilSubscriptions(t *testing.T) {
	var c Check
	b, err := json.Marshal(&c)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &c))
	require.NotNil(t, c.Subscriptions)
}

func TestCheckHasNonNilHandlers(t *testing.T) {
	var c Check
	b, err := json.Marshal(&c)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &c))
	require.NotNil(t, c.Handlers)
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

func TestCheckFlapThresholdValidation(t *testing.T) {
	c := FixtureCheck("foo")
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
