package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/token"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
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
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.Return(execution, nil)
	ch := make(chan *transport.Message, 5)
	agent.sendq = ch

	// check is already in progress, it shouldn't execute
	agent.inProgressMu.Lock()
	agent.inProgress[checkKey(request)] = request.Config
	agent.inProgressMu.Unlock()
	assert.Error(agent.handleCheck(context.TODO(), payload))

	// check is not in progress, it should execute
	agent.inProgressMu.Lock()
	delete(agent.inProgress, checkKey(request))
	agent.inProgressMu.Unlock()
	assert.NoError(agent.handleCheck(context.TODO(), payload))
}

func TestCheckInProgress_GH2704(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("normal-check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	key := checkKey(request)
	assert.Equal("normal-check", key)

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}

	agent.addInProgress(request)
	agent.inProgressMu.Lock()
	val, ok := agent.inProgress[key]
	agent.inProgressMu.Unlock()
	assert.True(ok)
	assert.True(agent.checkInProgress(request))
	assert.Equal(request.Config, val)

	agent.removeInProgress(request)
	agent.inProgressMu.Lock()
	val, ok = agent.inProgress[key]
	agent.inProgressMu.Unlock()
	assert.False(ok)
	assert.False(agent.checkInProgress(request))
	assert.Empty(val)
}

func TestProxyCheckInProgress_GH2704(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("proxy-check")
	checkConfig.ProxyEntityName = "proxy-entity"
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	key := checkKey(request)
	assert.Equal("proxy-check/proxy-entity", key)

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}

	agent.addInProgress(request)
	agent.inProgressMu.Lock()
	val, ok := agent.inProgress[key]
	agent.inProgressMu.Unlock()
	assert.True(ok)
	assert.True(agent.checkInProgress(request))
	assert.Equal(request.Config, val)

	agent.removeInProgress(request)
	agent.inProgressMu.Lock()
	val, ok = agent.inProgress[key]
	agent.inProgressMu.Unlock()
	assert.False(ok)
	assert.False(agent.checkInProgress(request))
	assert.Empty(val)
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

	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.Return(execution, nil)
	agent.sendq = make(chan *transport.Message, 5)

	// simulate checkA executing
	agent.inProgressMu.Lock()
	agent.inProgress[checkKey(reqA)] = reqA.Config
	agent.inProgressMu.Unlock()

	// check B should execute without error
	if err := agent.handleCheck(context.TODO(), payloadB); err != nil {
		t.Fatal(err)
	}

	// check A should not execute - in progress
	if err := agent.handleCheck(context.TODO(), payloadA); err == nil {
		t.Fatal("expected a non-nil error")
	}

	// After removing A from in-progress, it should execute without error
	agent.inProgressMu.Lock()
	delete(agent.inProgress, checkKey(reqA))
	agent.inProgressMu.Unlock()
	if err := agent.handleCheck(context.TODO(), payloadA); err != nil {
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
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.Return(execution, nil)

	entity := agent.getAgentEntity()
	agent.executeCheck(context.TODO(), request, entity)
	msg := <-ch

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(0), event.Check.Status)
	assert.False(event.HasMetrics())

	execution.Status = 1
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(1), event.Check.Status)
	assert.NotZero(event.Check.Issued)

	execution.Status = 127
	execution.Output = "command not found"
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(127), event.Check.Status)
	assert.Equal("command not found", event.Check.Output)
	assert.NotZero(event.Check.Issued)

	execution.Status = 2
	execution.Output = ""
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(2), event.Check.Status)

	checkConfig.OutputMetricHandlers = nil
	checkConfig.OutputMetricFormat = ""
	execution.Status = 0
	execution.Output = "metric.foo 1 123456789\nmetric.bar 2 987654321"
	ex.Return(execution, nil)
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.False(event.HasMetrics())

	checkConfig.OutputMetricFormat = corev2.GraphiteOutputMetricFormat
	ex.Return(execution, nil)
	agent.executeCheck(context.TODO(), request, entity)
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

func TestExecuteCheckDiscardOutput(t *testing.T) {
	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	output := "Here is some output"
	execution := command.FixtureExecutionResponse(0, output)
	ex.Return(execution, nil)

	agent.executeCheck(context.TODO(), request, agent.getAgentEntity())
	msg := <-ch

	event := &corev2.Event{}
	if err := json.Unmarshal(msg.Payload, event); err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Output, output; got != want {
		t.Fatal("check output incorrectly discarded")
	}

	request.Config.DiscardOutput = true

	agent.executeCheck(context.TODO(), request, agent.getAgentEntity())
	msg = <-ch

	if err := json.Unmarshal(msg.Payload, event); err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Output, ""; got != want {
		t.Fatal("check output not discarded")
	}
}

func TestHandleTokenSubstitution(t *testing.T) {
	assert := assert.New(t)

	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	config.AgentName = "TestTokenSubstitution"
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	// check command with valid token substitution
	checkConfig.Command = `echo {{ .name }} {{ .Missing | default "defaultValue" }}`
	checkConfig.Timeout = 10

	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	require.NoError(t, agent.handleCheck(context.TODO(), payload))

	msg := <-ch

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(0), event.Check.Status)
	assert.Contains(event.Check.Output, "TestTokenSubstitution defaultValue")
	assert.Contains(event.Check.Command, checkConfig.Command) // command should not include substitutions
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
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	// check command with unmatched token
	checkConfig.Command = `{{ .Foo }}`

	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	require.NoError(t, agent.handleCheck(context.TODO(), payload))

	msg := <-ch

	event := &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Contains(event.Check.Output, "has no entry for key")
	assert.Contains(event.Check.Command, checkConfig.Command)
}

func TestPrepareCheck(t *testing.T) {
	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}

	// Substitute
	entity := agent.getAgentEntity()
	entity.Labels = map[string]string{"foo": "bar"}
	check := corev2.FixtureCheckConfig("check")
	check.Command = "echo {{ .labels.foo }}"
	err = token.SubstituteCheck(check, entity)
	require.NoError(t, err)
	assert.Equal(t, check.Command, "echo bar")
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

func TestFailOnAssetCheckWithDisabledAssets(t *testing.T) {
	config, cleanup := FixtureConfig()
	defer cleanup()
	config.DisableAssets = true

	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	agent.sendq = make(chan *transport.Message, 5)
	checkConfig := corev2.FixtureCheckConfig("check")
	assets := []corev2.Asset{corev2.Asset{URL: "http://example.com/asset"}}
	request := &corev2.CheckRequest{Assets: assets, Config: checkConfig, Issued: time.Now().Unix()}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal("error marshaling check request")
	}
	if err := agent.handleCheck(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	msg := <-agent.sendq
	var event corev2.Event
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		t.Fatal(err)
	}
	if got, want := event.Check.Status, uint32(3); got != want {
		t.Errorf("bad status: got %d, want %d", got, want)
	}
}
