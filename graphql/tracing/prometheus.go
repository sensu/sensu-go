package tracing

import (
	"context"
	"fmt"

	time "github.com/echlebek/timeproxy"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/prometheus/client_golang/prometheus"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

const (
	KeyParse        = "parse"
	KeyValidate     = "validate"
	KeyExecuteQuery = "execute_query"
	KeyExecuteField = "execute_field"
)

var (
	Collector = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "graphql_duration_seconds",
			Help:       "Time spent in GraphQL operations, in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"key", "platform_key", "err"},
	)

	noopParse    = func(_ error) {}
	noopValidate = func(_ []gqlerrors.FormattedError) {}
	noopQuery    = func(_ *graphql.Result) {}
	noopField    = func(_ interface{}, _ error) {}

	platformKeys = map[string]string{
		KeyParse:        "graphql.parse",
		KeyValidate:     "graphql.validate",
		KeyExecuteQuery: "graphql.execute",
		KeyExecuteField: "graphql.execute",
	}
)

// PrometheusTracer is a GraphQL middleware that collects metrics for query
// parsing, validation, and field execution.
type PrometheusTracer struct {
	// AllowList contains the metrics that will be collected. Defaults to
	// []string{KeyExecuteField}
	AllowList []string
}

// NewPrometheusTracer instantiates new tracer.
func NewPrometheusTracer() *PrometheusTracer {
	return &PrometheusTracer{
		AllowList: []string{
			KeyExecuteField,
		},
	}
}

// Init is used to initialize the extension
func (t *PrometheusTracer) Init(ctx context.Context, p *graphql.Params) context.Context {
	return ctx
}

// Name returns the name of the extension
func (c *PrometheusTracer) Name() string {
	return "tracer.prometheus"
}

// ParseDidStart is called before starting parsing
func (c *PrometheusTracer) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	if !utilstrings.InArray(KeyParse, c.AllowList) {
		return ctx, noopParse
	}
	t := time.Now()
	return ctx, func(err error) {
		dur := msecSince(t)
		met := Collector.WithLabelValues(KeyParse, platformKeys[KeyParse], extractErrType(err))
		met.Observe(dur)
	}
}

// ValidationDidStart is called just before the validation begins
func (c *PrometheusTracer) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	if !utilstrings.InArray(KeyValidate, c.AllowList) {
		return ctx, noopValidate
	}
	t := time.Now()
	return ctx, func(errs []gqlerrors.FormattedError) {
		dur := msecSince(t)
		err := "<nil>"
		if len(errs) > 0 {
			err = extractErrType(errs[0].OriginalError())
		}
		met := Collector.WithLabelValues(KeyValidate, platformKeys[KeyValidate], err)
		met.Observe(dur)
	}
}

// ExecutionDidStart notifies about the start of the execution
func (c *PrometheusTracer) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	if !utilstrings.InArray(KeyExecuteQuery, c.AllowList) {
		return ctx, noopQuery
	}
	t := time.Now()
	return ctx, func(result *graphql.Result) {
		dur := msecSince(t)
		err := "<nil>"
		if result != nil && len(result.Errors) > 0 {
			err = extractErrType(result.Errors[0].OriginalError())
		}
		met := Collector.WithLabelValues(KeyExecuteQuery, platformKeys[KeyExecuteQuery], err)
		met.Observe(dur)
	}
}

// ResolveFieldDidStart notifies about the start of the resolving of a field
func (c *PrometheusTracer) ResolveFieldDidStart(ctx context.Context, i *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	if !utilstrings.InArray(KeyExecuteField, c.AllowList) {
		return ctx, noopField
	}
	t := time.Now()
	return ctx, func(_ interface{}, err error) {
		dur := msecSince(t)
		key := i.ParentType.Name() + "." + i.FieldName
		met := Collector.WithLabelValues(KeyExecuteField, key, extractErrType(err))
		met.Observe(dur)
	}
}

func (c *PrometheusTracer) GetResult(ctx context.Context) interface{} {
	return nil
}

func (c *PrometheusTracer) HasResult() bool {
	return false
}

func (c *PrometheusTracer) Collector() prometheus.Collector {
	return Collector
}

func msecSince(t time.Time) float64 {
	dur := time.Since(t)
	return float64(dur) / float64(time.Millisecond)
}

func extractErrType(err error) string {
	return fmt.Sprintf("%T", err)
}
