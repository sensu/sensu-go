package types

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

	// Invalid organization
	assert.Error(t, c.Validate())
	c.Organization = "default"

	// Invalid environment
	assert.Error(t, c.Validate())
	c.Environment = "default"

	// Invalid ttl
	c.Ttl = 10
	assert.Error(t, c.Validate())

	// Invalid metric format
	c.MetricFormat = "foo"
	assert.Error(t, c.Validate())
	c.MetricFormat = ""

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

func TestExtendedAttributes(t *testing.T) {
	type getter interface {
		Get(string) (interface{}, error)
	}
	check := FixtureCheck("chekov")
	check.SetExtendedAttributes([]byte(`{"foo":{"bar":42,"baz":9001}}`))
	g, err := check.Get("foo")
	require.NoError(t, err)
	v, err := g.(getter).Get("bar")
	require.NoError(t, err)
	require.Equal(t, 42.0, v)
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

	// Invalid entity attributes
	p.EntityAttributes = []string{`entity.Class = "proxy"`}
	assert.Error(t, p.Validate())
	p.EntityAttributes = []string{`entity.Class == "proxy"`}

	// Valid proxy request
	assert.NoError(t, p.Validate())
}

func TestMetricFormatValidate(t *testing.T) {
	assert.NoError(t, ValidateMetricFormat("nagios_perfdata"))
	assert.NoError(t, ValidateMetricFormat(NagiosMetricFormat))
	assert.NoError(t, ValidateMetricFormat(GraphiteMetricFormat))
	assert.NoError(t, ValidateMetricFormat(InfluxDBMetricFormat))
	assert.NoError(t, ValidateMetricFormat(OpenTSDBMetricFormat))
	assert.Error(t, ValidateMetricFormat("anything_else"))
	assert.Error(t, ValidateMetricFormat("NAGIOS_PERFDATA"))
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
