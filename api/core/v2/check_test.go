package v2

import (
	"encoding/json"
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

	c.Publish = false
	assert.NoError(t, c.Validate())
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

func TestOutputMetricFormatValidate(t *testing.T) {
	assert.NoError(t, ValidateOutputMetricFormat("nagios_perfdata"))
	assert.NoError(t, ValidateOutputMetricFormat(NagiosOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(GraphiteOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(InfluxDBOutputMetricFormat))
	assert.NoError(t, ValidateOutputMetricFormat(OpenTSDBOutputMetricFormat))
	assert.Error(t, ValidateOutputMetricFormat("anything_else"))
	assert.Error(t, ValidateOutputMetricFormat("NAGIOS_PERFDATA"))
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

func TestCheckMerge(t *testing.T) {
	a := FixtureCheck("check")
	b := FixtureCheck("check")

	for i := range a.History {
		if i%2 == 0 {
			a.History[i].Status = 1
		}
	}

	a.Occurrences = 1
	a.OccurrencesWatermark = 1

	b.History = nil
	b.TotalStateChange = 0
	b.State = ""
	b.Occurrences = 0
	b.OccurrencesWatermark = 0

	b.MergeWith(a)

	if got, want := b.TotalStateChange, uint32(98); got != want {
		t.Errorf("bad TotalStateChange: got %d, want %d", got, want)
	}

	if got, want := b.State, EventFlappingState; got != want {
		t.Errorf("bad State: got %s, want %s", got, want)
	}

	if got, want := b.Occurrences, int64(1); got != want {
		t.Errorf("bad Occurrences: got %d, want %d", got, want)
	}

	if got, want := b.OccurrencesWatermark, int64(1); got != want {
		t.Errorf("bad OccurrencesWatermark: got %d, want %d", got, want)
	}

}
