package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

var binDir = filepath.Join("..", "bin")
var toolsDir = filepath.Join(binDir, "tools")

func TestHandleCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath

	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 5)
	agent.sendq = ch

	// check is already in progress, it shouldn't execute
	agent.inProgressMu.Lock()
	agent.inProgress[request.Config.Name] = request.Config
	agent.inProgressMu.Unlock()
	assert.Error(agent.handleCheck(payload))

	// check is not in progress, it should execute
	agent.inProgressMu.Lock()
	delete(agent.inProgress, request.Config.Name)
	agent.inProgressMu.Unlock()
	assert.NoError(agent.handleCheck(payload))
}

func TestExecuteCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath
	checkConfig.Timeout = 10

	agent.executeCheck(request)

	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(0), event.Check.Status)

	falsePath := testutil.CommandPath(filepath.Join(toolsDir, "false"))
	checkConfig.Command = falsePath

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(1), event.Check.Status)
	assert.NotZero(event.Check.Issued)

	sleepPath := testutil.CommandPath(filepath.Join(toolsDir, "sleep"), "5")
	checkConfig.Command = sleepPath
	checkConfig.Timeout = 1

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(2), event.Check.Status)
}

func TestPrepareCheck(t *testing.T) {
	assert := assert.New(t)

	config := FixtureConfig()
	agent := NewAgent(config)

	// Invalid check
	check := types.FixtureCheckConfig("check")
	check.Interval = 0
	assert.False(agent.prepareCheck(check))

	// Valid check
	check.Interval = 60
	assert.True(agent.prepareCheck(check))
}

func TestExtractMetrics(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name            string
		event           *types.Event
		metricFormat    string
		expectedMetrics []*types.MetricPoint
	}{
		{
			name: "invalid metric format",
			event: &types.Event{
				Check: &types.Check{
					Output: "metric.value 1 123456789",
				},
			},
			metricFormat:    "not_a_format",
			expectedMetrics: nil,
		},
		{
			name: "valid extraction graphite",
			event: &types.Event{
				Check: &types.Check{
					Output: "metric.value 1 123456789",
				},
			},
			metricFormat: types.GraphiteMetricFormat,
			expectedMetrics: []*types.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			name: "invalid extraction graphite",
			event: &types.Event{
				Check: &types.Check{
					Output: "metric.value 1 foo",
				},
			},
			metricFormat:    types.GraphiteMetricFormat,
			expectedMetrics: nil,
		},
		{
			name: "valid nagios extraction",
			event: &types.Event{
				Check: &types.Check{
					Executed: 123456789,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0",
				},
			},
			metricFormat: types.NagiosMetricFormat,
			expectedMetrics: []*types.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*types.MetricTag{},
				},
			},
		},
		{
			name: "invalid nagios extraction",
			event: &types.Event{
				Check: &types.Check{
					Output: "PING ok - Packet loss = 0%",
				},
			},
			metricFormat:    types.NagiosMetricFormat,
			expectedMetrics: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.event.Check.MetricFormat = tc.metricFormat
			metrics := extractMetrics(tc.event)
			assert.Equal(tc.expectedMetrics, metrics)
		})
	}
}
