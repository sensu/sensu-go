package tracing

import (
	"context"
	"strings"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const summaryMetadata = `
  # HELP graphql_duration_seconds Time spent in GraphQL operations, in seconds
  # TYPE graphql_duration_seconds summary
`

func TestPrometheusTracer_ParseDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	tests := []struct {
		name      string
		allowList []string
		delta     time.Duration
		runs      int
		err       error
		want      string
	}{
		{
			name:      "omitted from allow list",
			allowList: []string{KeyValidate},
			delta:     10,
			runs:      5,
			want:      "",
		},
		{
			name:      "single run",
			allowList: []string{KeyParse},
			delta:     200,
			runs:      1,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.5"} 200
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.9"} 200
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.99"} 200
        graphql_duration_seconds_sum{err="<nil>",key="parse",platform_key="graphql.parse"} 200
        graphql_duration_seconds_count{err="<nil>",key="parse",platform_key="graphql.parse"} 1
      `,
		},
		{
			name:      "multiple runs",
			allowList: []string{KeyParse},
			delta:     20,
			runs:      5,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.5"} 20
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.9"} 20
        graphql_duration_seconds{err="<nil>",key="parse",platform_key="graphql.parse",quantile="0.99"} 20
        graphql_duration_seconds_sum{err="<nil>",key="parse",platform_key="graphql.parse"} 100
        graphql_duration_seconds_count{err="<nil>",key="parse",platform_key="graphql.parse"} 5
      `,
		},
		{
			name:      "encountered error",
			allowList: []string{KeyParse},
			delta:     20,
			runs:      5,
			err:       context.DeadlineExceeded,
			want: summaryMetadata + `
        graphql_duration_seconds{err="context.deadlineExceededError",key="parse",platform_key="graphql.parse",quantile="0.5"} 20
        graphql_duration_seconds{err="context.deadlineExceededError",key="parse",platform_key="graphql.parse",quantile="0.9"} 20
        graphql_duration_seconds{err="context.deadlineExceededError",key="parse",platform_key="graphql.parse",quantile="0.99"} 20
        graphql_duration_seconds_sum{err="context.deadlineExceededError",key="parse",platform_key="graphql.parse"} 100
        graphql_duration_seconds_count{err="context.deadlineExceededError",key="parse",platform_key="graphql.parse"} 5
      `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := newTracer(tt.allowList)

			for i := 0; i < tt.runs; i++ {
				_, fn := trace.ParseDidStart(context.Background())
				mockTime.Set(time.Now().Add(tt.delta * time.Millisecond))
				fn(tt.err)
			}

			if err := testutil.CollectAndCompare(trace.Collector(), strings.NewReader(tt.want), "graphql_duration_seconds"); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}
		})
	}
}

func TestPrometheusTracer_ValidationDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	tests := []struct {
		name      string
		allowList []string
		delta     time.Duration
		err       error
		runs      int
		want      string
	}{
		{
			name:      "omitted from allow list",
			allowList: []string{KeyParse},
			delta:     10,
			runs:      50,
			want:      "",
		},
		{
			name:      "single run",
			allowList: []string{KeyValidate},
			delta:     150,
			runs:      1,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.5"} 150
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.9"} 150
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.99"} 150
        graphql_duration_seconds_sum{err="<nil>",key="validate",platform_key="graphql.validate"} 150
        graphql_duration_seconds_count{err="<nil>",key="validate",platform_key="graphql.validate"} 1
      `,
		},
		{
			name:      "multiple runs",
			allowList: []string{KeyValidate},
			delta:     15,
			runs:      8,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.5"} 15
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.9"} 15
        graphql_duration_seconds{err="<nil>",key="validate",platform_key="graphql.validate",quantile="0.99"} 15
        graphql_duration_seconds_sum{err="<nil>",key="validate",platform_key="graphql.validate"} 120
        graphql_duration_seconds_count{err="<nil>",key="validate",platform_key="graphql.validate"} 8
      `,
		},
		{
			name:      "encountered error",
			allowList: []string{KeyValidate},
			delta:     15,
			runs:      8,
			err:       context.DeadlineExceeded,
			want: summaryMetadata + `
        graphql_duration_seconds{err="context.deadlineExceededError",key="validate",platform_key="graphql.validate",quantile="0.5"} 15
        graphql_duration_seconds{err="context.deadlineExceededError",key="validate",platform_key="graphql.validate",quantile="0.9"} 15
        graphql_duration_seconds{err="context.deadlineExceededError",key="validate",platform_key="graphql.validate",quantile="0.99"} 15
        graphql_duration_seconds_sum{err="context.deadlineExceededError",key="validate",platform_key="graphql.validate"} 120
        graphql_duration_seconds_count{err="context.deadlineExceededError",key="validate",platform_key="graphql.validate"} 8
      `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := newTracer(tt.allowList)

			for i := 0; i < tt.runs; i++ {
				_, fn := trace.ValidationDidStart(context.Background())
				mockTime.Set(time.Now().Add(tt.delta * time.Millisecond))

				var errs []gqlerrors.FormattedError
				if tt.err != nil {
					errs = gqlerrors.FormatErrors(tt.err)
				}
				fn(errs)
			}

			if err := testutil.CollectAndCompare(trace.Collector(), strings.NewReader(tt.want), "graphql_duration_seconds"); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}
		})
	}
}

func TestPrometheusTracer_ExecutionDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	tests := []struct {
		name      string
		allowList []string
		delta     time.Duration
		runs      int
		err       error
		want      string
	}{
		{
			name:      "omitted from allow list",
			allowList: []string{KeyParse, KeyValidate},
			delta:     10,
			runs:      50,
			want:      "",
		},
		{
			name:      "single run",
			allowList: []string{KeyExecuteQuery},
			delta:     120,
			runs:      1,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.5"} 120
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.9"} 120
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.99"} 120
        graphql_duration_seconds_sum{err="<nil>",key="execute_query",platform_key="graphql.execute"} 120
        graphql_duration_seconds_count{err="<nil>",key="execute_query",platform_key="graphql.execute"} 1
      `,
		},
		{
			name:      "multiple runs",
			allowList: []string{KeyExecuteQuery},
			delta:     12,
			runs:      9,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.5"} 12
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.9"} 12
        graphql_duration_seconds{err="<nil>",key="execute_query",platform_key="graphql.execute",quantile="0.99"} 12
        graphql_duration_seconds_sum{err="<nil>",key="execute_query",platform_key="graphql.execute"} 108
        graphql_duration_seconds_count{err="<nil>",key="execute_query",platform_key="graphql.execute"} 9
      `,
		}, {
			name:      "encountered error",
			allowList: []string{KeyExecuteQuery},
			delta:     12,
			runs:      9,
			err:       context.DeadlineExceeded,
			want: summaryMetadata + `
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_query",platform_key="graphql.execute",quantile="0.5"} 12
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_query",platform_key="graphql.execute",quantile="0.9"} 12
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_query",platform_key="graphql.execute",quantile="0.99"} 12
        graphql_duration_seconds_sum{err="context.deadlineExceededError",key="execute_query",platform_key="graphql.execute"} 108
        graphql_duration_seconds_count{err="context.deadlineExceededError",key="execute_query",platform_key="graphql.execute"} 9
      `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := newTracer(tt.allowList)

			for i := 0; i < tt.runs; i++ {
				_, fn := trace.ExecutionDidStart(context.Background())
				mockTime.Set(time.Now().Add(tt.delta * time.Millisecond))

				result := &graphql.Result{}
				if tt.err != nil {
					result.Errors = gqlerrors.FormatErrors(tt.err)
				}
				fn(result)
			}

			if err := testutil.CollectAndCompare(trace.Collector(), strings.NewReader(tt.want), "graphql_duration_seconds"); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}
		})
	}
}

func TestPrometheusTracer_ResolveFieldDidStart(t *testing.T) {
	mockTime, cleanup := mockTime()
	defer cleanup()

	tests := []struct {
		name      string
		allowList []string
		delta     time.Duration
		runs      int
		err       error
		want      string
	}{
		{
			name:      "omitted from allow list",
			allowList: []string{KeyExecuteQuery},
			delta:     10,
			runs:      20,
			want:      "",
		},
		{
			name:      "single run",
			allowList: []string{KeyExecuteField},
			delta:     120,
			runs:      1,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.5"} 120
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.9"} 120
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.99"} 120
        graphql_duration_seconds_sum{err="<nil>",key="execute_field",platform_key="Check.interval"} 120
        graphql_duration_seconds_count{err="<nil>",key="execute_field",platform_key="Check.interval"} 1
      `,
		},
		{
			name:      "multiple runs",
			allowList: []string{KeyExecuteField},
			delta:     15,
			runs:      7,
			want: summaryMetadata + `
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.5"} 15
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.9"} 15
        graphql_duration_seconds{err="<nil>",key="execute_field",platform_key="Check.interval",quantile="0.99"} 15
        graphql_duration_seconds_sum{err="<nil>",key="execute_field",platform_key="Check.interval"} 105
        graphql_duration_seconds_count{err="<nil>",key="execute_field",platform_key="Check.interval"} 7
      `,
		},
		{
			name:      "encountered error",
			allowList: []string{KeyExecuteField},
			delta:     15,
			runs:      7,
			err:       context.DeadlineExceeded,
			want: summaryMetadata + `
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_field",platform_key="Check.interval",quantile="0.5"} 15
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_field",platform_key="Check.interval",quantile="0.9"} 15
        graphql_duration_seconds{err="context.deadlineExceededError",key="execute_field",platform_key="Check.interval",quantile="0.99"} 15
        graphql_duration_seconds_sum{err="context.deadlineExceededError",key="execute_field",platform_key="Check.interval"} 105
        graphql_duration_seconds_count{err="context.deadlineExceededError",key="execute_field",platform_key="Check.interval"} 7
      `,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trace := newTracer(tt.allowList)
			info := &graphql.ResolveInfo{
				FieldName:  "interval",
				ParentType: &checkType{},
			}

			for i := 0; i < tt.runs; i++ {
				_, fn := trace.ResolveFieldDidStart(context.Background(), info)
				mockTime.Set(time.Now().Add(tt.delta * time.Millisecond))
				fn(nil, tt.err)
			}

			if err := testutil.CollectAndCompare(trace.Collector(), strings.NewReader(tt.want), "graphql_duration_seconds"); err != nil {
				t.Errorf("unexpected collecting result:\n%s", err)
			}
		})
	}
}

func mockTime() (*crock.Time, func()) {
	mockTime := crock.NewTime(time.Unix(0, 0))
	time.TimeProxy = mockTime
	return mockTime, func() {
		time.TimeProxy = time.RealTime{}
	}
}

func newTracer(allowList []string) *PrometheusTracer {
	Collector.Reset() // prevent metrics from leaking between tests
	trace := NewPrometheusTracer()
	trace.AllowList = allowList
	return trace
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
