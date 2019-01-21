package agent

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("check")

	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent := NewAgent(config)
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
	ch := make(chan *transport.Message, 5)
	agent.sendq = ch

	// check is already in progress, it shouldn't execute
	agent.inProgressMu.Lock()
	agent.inProgress[checkKey(request)] = request.Config
	agent.inProgressMu.Unlock()
	assert.Error(agent.handleCheck(payload))

	// check is not in progress, it should execute
	agent.inProgressMu.Lock()
	delete(agent.inProgress, checkKey(request))
	agent.inProgressMu.Unlock()
	assert.NoError(agent.handleCheck(payload))
}

func TestHandleProxyCheck(t *testing.T) {
	checkA := corev2.FixtureCheckConfig("check")
	checkA.ProxyEntityName = "A"

	checkB := corev2.FixtureCheckConfig("check")
	checkB.ProxyEntityName = "B"

	reqA := &corev2.CheckRequest{Config: checkA, Issued: time.Now().Unix()}
	reqB := &corev2.CheckRequest{Config: checkB, Issued: time.Now().Unix()}

	payloadA, _ := json.Marshal(reqA)
	payloadB, _ := json.Marshal(reqB)

	config, cleanup := FixtureConfig()
	defer cleanup()

	agent := NewAgent(config)
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
	agent.sendq = make(chan *transport.Message, 5)

	// simulate checkA executing
	agent.inProgressMu.Lock()
	agent.inProgress[checkKey(reqA)] = reqA.Config
	agent.inProgressMu.Unlock()

	// check B should execute without error
	if err := agent.handleCheck(payloadB); err != nil {
		t.Fatal(err)
	}

	// check A should not execute - in progress
	if err := agent.handleCheck(payloadA); err == nil {
		t.Fatal("expected a non-nil error")
	}

	// After removing A from in-progress, it should execute without error
	agent.inProgressMu.Lock()
	delete(agent.inProgress, checkKey(reqA))
	agent.inProgressMu.Unlock()
	if err := agent.handleCheck(payloadA); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)

	agent.executeCheck(request)
	msg := <-ch

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(0), event.Check.Status)
	assert.False(event.HasMetrics())

	execution.Status = 1
	agent.executeCheck(request)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(1), event.Check.Status)
	assert.NotZero(event.Check.Issued)

	execution.Status = 127
	execution.Output = "command not found"
	agent.executeCheck(request)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(127), event.Check.Status)
	assert.Equal("command not found", event.Check.Output)
	assert.NotZero(event.Check.Issued)

	execution.Status = 2
	execution.Output = ""
	agent.executeCheck(request)
	msg = <-ch

	event = &corev2.Event{}
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

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.False(event.HasMetrics())

	checkConfig.OutputMetricFormat = corev2.GraphiteOutputMetricFormat
	ex.On("Execute", mock.Anything, mock.Anything).Return(execution, nil)
	agent.executeCheck(request)
	msg = <-ch

	event = &corev2.Event{}
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

	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	config.AgentName = "TestTokenSubstitution"
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	// check command with valid token substitution
	checkConfig.Command = `echo {{ .name }} {{ .Missing | default "defaultValue" }}`
	checkConfig.Timeout = 10

	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	require.NoError(t, agent.handleCheck(payload))

	msg := <-ch

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(0), event.Check.Status)
	assert.Contains(event.Check.Output, "TestTokenSubstitution defaultValue")
}

func TestHandleTokenSubstitutionNoKey(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	config.Labels = map[string]string{"team": "devops"}
	config.AgentName = "TestTokenSubstitution"
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

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Contains(event.Check.Output, "has no entry for key")
}

func TestPrepareCheck(t *testing.T) {
	assert := assert.New(t)

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent := NewAgent(config)

	// Invalid check
	check := corev2.FixtureCheckConfig("check")
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
		event           *corev2.Event
		metricFormat    string
		expectedMetrics []*corev2.MetricPoint
	}{
		{
			name: "invalid output metric format",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "metric.value 1 123456789",
				},
			},
			metricFormat:    "not_a_format",
			expectedMetrics: nil,
		},
		{
			name: "valid extraction graphite",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "metric.value 1 123456789",
				},
			},
			metricFormat: corev2.GraphiteOutputMetricFormat,
			expectedMetrics: []*corev2.MetricPoint{
				{
					Name:      "metric.value",
					Value:     1,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			name: "invalid extraction graphite",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "metric.value 1 foo",
				},
			},
			metricFormat:    corev2.GraphiteOutputMetricFormat,
			expectedMetrics: nil,
		},
		{
			name: "valid nagios extraction",
			event: &corev2.Event{
				Check: &corev2.Check{
					Executed: 123456789,
					Output:   "PING ok - Packet loss = 0% | percent_packet_loss=0",
				},
			},
			metricFormat: corev2.NagiosOutputMetricFormat,
			expectedMetrics: []*corev2.MetricPoint{
				{
					Name:      "percent_packet_loss",
					Value:     0,
					Timestamp: 123456789,
					Tags:      []*corev2.MetricTag{},
				},
			},
		},
		{
			name: "invalid nagios extraction",
			event: &corev2.Event{
				Check: &corev2.Check{
					Output: "PING ok - Packet loss = 0%",
				},
			},
			metricFormat:    corev2.NagiosOutputMetricFormat,
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
