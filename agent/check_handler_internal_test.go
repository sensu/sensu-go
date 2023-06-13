package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"
	"github.com/sensu/sensu-go/token"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCheck(t *testing.T) {
	newCheckRequest := func(t *testing.T, issued int64) []byte {
		checkConfig := corev2.FixtureCheckConfig("check")
		checkRequest := &corev2.CheckRequest{
			Config: checkConfig,
			Issued: issued,
		}
		payload, err := json.Marshal(checkRequest)
		require.NoError(t, err)
		return payload
	}
	type fields struct {
		inProgress map[string]*corev2.CheckConfig
		lastIssued map[string]int64
		sequences  map[string]int64
	}
	type args struct {
		ctx     context.Context
		payload []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "errors when issued timestamp is older than the previous",
			fields: fields{
				lastIssued: map[string]int64{
					"check": time.Now().Unix() + 1000,
				},
			},
			args: args{
				ctx:     context.Background(),
				payload: newCheckRequest(t, time.Now().Unix()),
			},
			wantErr:    true,
			wantErrMsg: errOldCheckRequest.Error(),
		},
		{
			name: "errors when issued timestamp is the same as the previous",
			fields: fields{
				lastIssued: map[string]int64{
					"check": 1234,
				},
			},
			args: args{
				ctx:     context.Background(),
				payload: newCheckRequest(t, 1234),
			},
			wantErr:    true,
			wantErrMsg: errDupCheckRequest.Error(),
		},
		{
			name: "errors when check execution is already in progress",
			fields: fields{
				inProgress: map[string]*corev2.CheckConfig{
					"check": corev2.FixtureCheckConfig("check"),
				},
				lastIssued: map[string]int64{
					"check": 1234,
				},
			},
			args: args{
				ctx:     context.Background(),
				payload: newCheckRequest(t, 1235),
			},
			wantErr:    true,
			wantErrMsg: "check execution still in progress: check",
		},
		{
			name: "executes the first time a request for a check is received",
			fields: fields{
				inProgress: make(map[string]*corev2.CheckConfig),
				lastIssued: make(map[string]int64),
				sequences:  make(map[string]int64),
			},
			args: args{
				ctx:     context.Background(),
				payload: newCheckRequest(t, 1235),
			},
			wantErr: false,
		},
		{
			name: "executes the second time a request for a check is received",
			fields: fields{
				inProgress: make(map[string]*corev2.CheckConfig),
				lastIssued: map[string]int64{
					"check": 1234,
				},
				sequences: map[string]int64{
					"check": 1,
				},
			},
			args: args{
				ctx:     context.Background(),
				payload: newCheckRequest(t, 1235),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentConfig := &Config{}
			ex := &mockexecutor.MockExecutor{}
			execution := command.FixtureExecutionResponse(0, "")
			ex.Return(execution, nil)
			agent := &Agent{
				config:       agentConfig,
				executor:     ex,
				inProgress:   tt.fields.inProgress,
				inProgressMu: &sync.Mutex{},
				lastIssued:   tt.fields.lastIssued,
				lastIssuedMu: &sync.Mutex{},
				marshal:      MarshalJSON,
				sequences:    tt.fields.sequences,
				unmarshal:    UnmarshalJSON,
			}
			err := agent.handleCheck(tt.args.ctx, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.handleCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("Agent.handleCheck() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
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
	assert.Equal(event.Sequence, int64(1))

	execution.Status = 1
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(1), event.Check.Status)
	assert.NotZero(event.Check.Issued)
	assert.Equal(event.Sequence, int64(2))

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
	assert.Equal(event.Sequence, int64(3))

	execution.Status = 2
	execution.Output = ""
	agent.executeCheck(context.TODO(), request, entity)
	msg = <-ch

	event = &corev2.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.Equal(uint32(2), event.Check.Status)
	assert.Equal(event.Sequence, int64(4))

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
	assert.Equal(event.Sequence, int64(5))

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
	assert.Equal(event.Sequence, int64(6))
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
	assert.NotZero(event.Sequence)
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

func TestCheckHandlerProcessedBy(t *testing.T) {
	checkConfig := corev2.FixtureCheckConfig("check")
	request := &corev2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	config.AgentName = "boris"
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

	if got, want := event.Check.ProcessedBy, "boris"; got != want {
		t.Errorf("bad processed_by: got %q, want %q", got, want)
	}
}

func TestEvaluateOutputMetricThresholds(t *testing.T) {
	now := time.Now().UnixMilli()

	metric1 := &corev2.MetricPoint{Name: "disk_rate", Value: 99999.0, Timestamp: now, Tags: nil}
	metric2 := &corev2.MetricPoint{Name: "network_rate", Value: 100001.0, Timestamp: now, Tags: []*corev2.MetricTag{{Name: "device", Value: "eth0"}}}

	statusOKAnnotation := "sensu.io/notifications/ok"
	statusWarningAnnotation := "sensu.io/notifications/warning"
	statusUnknownAnnotation := "sensu.io/notifications/unknown"
	statusCriticalAnnotation := "sensu.io/notifications/critical"
	diskOKAnnotation := "sensu.io/output_metric_thresholds/disk_rate/ok"
	diskCriticalAnnotation := "sensu.io/output_metric_thresholds/disk_rate/critical"
	diskWarningAnnotation := "sensu.io/output_metric_thresholds/disk_rate/warning"
	netUnknownAnnotation := "sensu.io/output_metric_thresholds/network_rate/unknown"
	notDiskWarningNullAnnotation := "sensu.io/output_metric_thresholds/not_a_disk_rate/warning"

	testCases := []struct {
		name                string
		event               *corev2.Event
		metrics             []*corev2.MetricPoint
		thresholds          []*corev2.MetricThreshold
		expectedStatus      uint32
		expectedAnnotations []string
	}{
		{
			name:                "minimum rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Min: "200000.0", Status: 2}}}},
			expectedStatus:      2,
			expectedAnnotations: []string{statusCriticalAnnotation, diskCriticalAnnotation},
		}, {
			name:                "maximum rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Max: "50000.0", Status: 2}}}},
			expectedStatus:      2,
			expectedAnnotations: []string{statusCriticalAnnotation, diskCriticalAnnotation},
		}, {
			name:                "no min rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Min: "50000.0", Status: 2}}}},
			expectedStatus:      0,
			expectedAnnotations: []string{statusOKAnnotation, diskOKAnnotation},
		}, {
			name:                "no max rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Max: "200000.0", Status: 2}}}},
			expectedStatus:      0,
			expectedAnnotations: []string{statusOKAnnotation, diskOKAnnotation},
		}, {
			name:                "min and max rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Min: "200000.0", Status: 1}, {Max: "75000.0", Status: 2}}}},
			expectedStatus:      2,
			expectedAnnotations: []string{statusCriticalAnnotation, diskCriticalAnnotation},
		}, {
			name:                "only one rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", Thresholds: []*corev2.MetricThresholdRule{{Min: "200000.0", Status: 1}, {Max: "200000.0", Status: 2}}}},
			expectedStatus:      1,
			expectedAnnotations: []string{statusWarningAnnotation, diskWarningAnnotation},
		}, {
			name:                "no filter match - null status",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1},
			thresholds:          []*corev2.MetricThreshold{{Name: "not_a_disk_rate", NullStatus: 1, Thresholds: []*corev2.MetricThresholdRule{{Max: "200000.0", Status: 2}}}},
			expectedStatus:      1,
			expectedAnnotations: []string{statusWarningAnnotation, notDiskWarningNullAnnotation},
		}, {
			name:                "multi metric and filter match, no rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1, metric2},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", NullStatus: 1, Thresholds: []*corev2.MetricThresholdRule{{Max: "200000.0", Status: 2}}}},
			expectedStatus:      0,
			expectedAnnotations: []string{statusOKAnnotation, diskOKAnnotation},
		}, {
			name:                "multi metric and filter and rule match",
			event:               &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics:             []*corev2.MetricPoint{metric1, metric2},
			thresholds:          []*corev2.MetricThreshold{{Name: "disk_rate", NullStatus: 1, Thresholds: []*corev2.MetricThresholdRule{{Max: "50000.0", Status: 2}}}},
			expectedStatus:      2,
			expectedAnnotations: []string{statusCriticalAnnotation, diskCriticalAnnotation},
		}, {
			name:    "multi metric and multi rule match",
			event:   &corev2.Event{Check: &corev2.Check{Status: 0}},
			metrics: []*corev2.MetricPoint{metric1, metric2},
			thresholds: []*corev2.MetricThreshold{{Name: "disk_rate", NullStatus: 1, Thresholds: []*corev2.MetricThresholdRule{{Max: "50000.0", Status: 2}}},
				{Name: "network_rate", Thresholds: []*corev2.MetricThresholdRule{{Max: "40000", Status: 3}}}},
			expectedStatus:      3,
			expectedAnnotations: []string{statusUnknownAnnotation, diskCriticalAnnotation, netUnknownAnnotation},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			event := test.event
			test.event.Metrics = &corev2.Metrics{Points: test.metrics}
			test.event.Check.OutputMetricThresholds = test.thresholds
			status := evaluateOutputMetricThresholds(event)
			assert.Equal(t, test.expectedStatus, status)

			assert.Equal(t, len(test.expectedAnnotations), len(event.Annotations), "wrong annotation count")
			for _, expectedKey := range test.expectedAnnotations {
				_, ok := event.Annotations[expectedKey]
				assert.True(t, ok, fmt.Sprintf("missing annotation %s", expectedKey))
			}
		})
	}
}
