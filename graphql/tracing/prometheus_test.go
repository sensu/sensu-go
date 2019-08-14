package tracing

import (
	"context"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	"github.com/gogo/protobuf/proto"
	"github.com/graphql-go/graphql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusTracerParseDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	trace := NewPrometheusTracer()
	trace.AllowList = []string{KeyParse}

	_, fn := trace.ParseDidStart(context.Background())
	mockTime.Set(time.Now().Add(7 * time.Millisecond))
	fn(nil)

	got := mustGather(t, trace.Collector())
	assert.Contains(t, got, `
  label: <
    name: "key"
    value: "parse"
  >
  label: <
    name: "platform_key"
    value: "graphql.parse"
  >
  summary: <
    sample_count: 1
    sample_sum: 7
    quantile: <
      quantile: 0.5
      value: 7
    >
    quantile: <
      quantile: 0.9
      value: 7
    >
    quantile: <
      quantile: 0.99
      value: 7
    >
  >`)

	trace.AllowList = []string{}
	_, fn = trace.ParseDidStart(context.Background())
	fn(nil)

	got = mustGather(t, trace.Collector())
	assert.Contains(t, got, "sample_count: 1")
}

func TestPrometheusTracerValidateDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	trace := NewPrometheusTracer()
	trace.AllowList = []string{KeyValidate}

	_, fn := trace.ValidationDidStart(context.Background())
	mockTime.Set(time.Now().Add(540 * time.Millisecond))
	fn(nil)

	got := mustGather(t, trace.Collector())
	assert.Contains(t, got, `
  label: <
    name: "key"
    value: "validate"
  >
  label: <
    name: "platform_key"
    value: "graphql.validate"
  >
  summary: <
    sample_count: 1
    sample_sum: 540
    quantile: <
      quantile: 0.5
      value: 540
    >
    quantile: <
      quantile: 0.9
      value: 540
    >
    quantile: <
      quantile: 0.99
      value: 540
    >
  >`)

	trace.AllowList = []string{}
	_, fn = trace.ValidationDidStart(context.Background())
	fn(nil)

	got = mustGather(t, trace.Collector())
	assert.Contains(t, got, "sample_count: 1")
}

func TestPrometheusTracerExecutionDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	trace := NewPrometheusTracer()
	trace.AllowList = []string{KeyExecuteQuery}

	_, fn := trace.ExecutionDidStart(context.Background())
	mockTime.Set(time.Now().Add(720 * time.Millisecond))
	fn(nil)

	got := mustGather(t, trace.Collector())
	assert.Contains(t, got, `
  label: <
    name: "key"
    value: "execute_query"
  >
  label: <
    name: "platform_key"
    value: "graphql.execute"
  >
  summary: <
    sample_count: 1
    sample_sum: 720
    quantile: <
      quantile: 0.5
      value: 720
    >
    quantile: <
      quantile: 0.9
      value: 720
    >
    quantile: <
      quantile: 0.99
      value: 720
    >
  >`)

	trace.AllowList = []string{}
	_, fn = trace.ExecutionDidStart(context.Background())
	fn(nil)

	got = mustGather(t, trace.Collector())
	assert.Contains(t, got, "sample_count: 1")
}

func TestPrometheusTracerFieldDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	trace := NewPrometheusTracer()
	trace.AllowList = []string{KeyExecuteField}

	info := &graphql.ResolveInfo{
		FieldName:  "interval",
		ParentType: &checkType{},
	}

	_, fn := trace.ResolveFieldDidStart(context.Background(), info)
	mockTime.Set(time.Now().Add(1530 * time.Millisecond))
	fn(nil, nil)

	got := mustGather(t, trace.Collector())
	assert.Contains(t, got, `
  label: <
    name: "key"
    value: "execute_field"
  >
  label: <
    name: "platform_key"
    value: "Check.interval"
  >
  summary: <
    sample_count: 1
    sample_sum: 1530
    quantile: <
      quantile: 0.5
      value: 1530
    >
    quantile: <
      quantile: 0.9
      value: 1530
    >
    quantile: <
      quantile: 0.99
      value: 1530
    >
  >`)

	trace.AllowList = []string{}
	_, fn = trace.ResolveFieldDidStart(context.Background(), info)
	fn(nil, nil)

	got = mustGather(t, trace.Collector())
	assert.Contains(t, got, "sample_count: 1")
}

func mockTime() (*crock.Time, func()) {
	mockTime := crock.NewTime(time.Unix(0, 0))
	time.TimeProxy = mockTime
	return mockTime, func() {
		time.TimeProxy = time.RealTime{}
	}
}

func mustGather(t *testing.T, collector prometheus.Collector) string {
	reg := prometheus.NewRegistry()
	err := reg.Register(collector)
	if err != nil {
		t.Error("unable to register tracer")
		t.FailNow()
	}
	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Errorf("unexpected behavior of custom test registry.\n%#v", metricFamilies)
		t.FailNow()
	}
	if len(metricFamilies) == 0 {
		return ""
	}
	return proto.MarshalTextString(metricFamilies[0])
}

type checkType struct{}

func (*checkType) Name() string {
	return "Check"
}
func (*checkType) Description() string {
	return ""
}
func (*checkType) String() string {
	return ""
}
func (*checkType) Error() error {
	return nil
}
