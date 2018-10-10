package agent

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")

	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	config := FixtureConfig()
	agent := NewAgent(config)
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
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
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)

	agent.executeCheck(request)
	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(0), event.Check.Status)
	assert.False(event.HasMetrics())

	execution.Status = 1
	agent.executeCheck(request)
	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(1), event.Check.Status)
	assert.NotZero(event.Check.Issued)

	execution.Status = 127
	execution.Output = "command not found"
	agent.executeCheck(request)
	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(127), event.Check.Status)
	assert.Equal("command not found", event.Check.Output)
	assert.NotZero(event.Check.Issued)

	execution.Status = 2
	execution.Output = ""
	agent.executeCheck(request)
	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(2), event.Check.Status)

	checkConfig.OutputMetricHandlers = nil
	checkConfig.OutputMetricFormat = ""
	execution.Status = 0
	execution.Output = "metric.foo 1 123456789\nmetric.bar 2 987654321"
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
	agent.executeCheck(request)
	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.False(event.HasMetrics())

	checkConfig.OutputMetricFormat = types.GraphiteOutputMetricFormat
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
	agent.executeCheck(request)
	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.True(event.HasMetrics())
	require.Equal(t, 2, len(event.Metrics.Points), string(msg.Payload))
	metric0 := event.Metrics.Points[0]
	assert.Equal(float64(1), metric0.Value)
	assert.Equal("metric.foo", metric0.Name)
	assert.Equal(int64(123456789), metric0.Timestamp)
	metric1 := event.Metrics.Points[1]
	assert.Equal(float64(2), metric1.Value)
	assert.Equal("metric.bar", metric1.Name)
	assert.Equal(int64(987654321), metric1.Timestamp)
}

func TestHandleTokenSubstitution(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config := FixtureConfig()
	config.ExtendedAttributes = []byte(`{"team":"devops"}`)
	config.AgentID = "TestTokenSubstitution"
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	// check command with valid token substitution
	checkConfig.Command = `echo {{ .ID }} {{ .Team }} {{ .Missing | default "defaultValue" }}`
	checkConfig.Timeout = 10

	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	require.NoError(t, agent.handleCheck(payload))

	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(0), event.Check.Status)
	assert.Contains(event.Check.Output, "TestTokenSubstitution devops defaultValue")
}

func TestHandleTokenSubstitutionNoKey(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config := FixtureConfig()
	config.ExtendedAttributes = []byte(`{"team":"devops"}`)
	config.AgentID = "TestTokenSubstitution"
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	// check command with unmatched token
	checkConfig.Command = `{{ .Foo }}`

	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	require.NoError(t, agent.handleCheck(payload))

	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Contains(event.Check.Output, "has no entry for key")
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
			name: "invalid output metric format",
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
			metricFormat: types.GraphiteOutputMetricFormat,
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
			metricFormat:    types.GraphiteOutputMetricFormat,
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
			metricFormat: types.NagiosOutputMetricFormat,
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
			metricFormat:    types.NagiosOutputMetricFormat,
			expectedMetrics: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.event.Check.OutputMetricFormat = tc.metricFormat
			metrics := extractMetrics(tc.event)
			assert.Equal(tc.expectedMetrics, metrics)
		})
	}
}
