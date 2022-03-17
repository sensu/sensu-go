package graphql

import (
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/graphql"
)

var _ graphql.InterfaceTypeResolver = (*metricImpl)(nil)
var _ schema.CounterMetricFieldResolvers = (*counterMetricImpl)(nil)
var _ schema.GaugeMetricFieldResolvers = (*gaugeMetricImpl)(nil)
var _ schema.SummaryMetricFieldResolvers = (*summaryMetricImpl)(nil)
var _ schema.UntypedMetricFieldResolvers = (*untypedMetricImpl)(nil)
var _ schema.HistogramMetricFieldResolvers = (*histogramMetricImpl)(nil)
var _ schema.MetricFamilyFieldResolvers = (*metricFamilyImpl)(nil)

//
// Implement InterfaceTypeResolver for Metric
//

type metricImpl struct{}

func (*metricImpl) ResolveType(v interface{}, _ graphql.ResolveTypeParams) *graphql.Type {
	mf, ok := v.(*dto.Metric)
	if !ok {
		logger.Fatalf("Metric#ResolveType received unexpected type %T", v)
		return nil
	}
	switch {
	case mf.Counter != nil:
		return &schema.CounterMetricType
	case mf.Gauge != nil:
		return &schema.GaugeMetricType
	case mf.Summary != nil:
		return &schema.SummaryMetricType
	case mf.Untyped != nil:
		return &schema.UntypedMetricType
	case mf.Histogram != nil:
		return &schema.HistogramMetricType
	}
	logger.WithField("metric", mf).Fatal("Metric#ResolveType received unexpected prom metric")
	return nil
}

//
// Implement CounterMetricFieldResolvers
//

type counterMetricImpl struct {
	*commonMetricImpl
}

func (*counterMetricImpl) Value(p graphql.ResolveParams) (float64, error) {
	v := p.Source.(*dto.Metric)
	return formatFloat(v.Counter.Value), nil
}

//
// Implement GuageMetricFieldResolvers
//

type gaugeMetricImpl struct {
	*commonMetricImpl
}

func (*gaugeMetricImpl) Value(p graphql.ResolveParams) (float64, error) {
	v := p.Source.(*dto.Metric)
	return formatFloat(v.Gauge.Value), nil
}

//
// Implement SummaryMetricFieldResolvers
//

type summaryMetricImpl struct {
	*commonMetricImpl
}

func (*summaryMetricImpl) SampleCount(p graphql.ResolveParams) (int, error) {
	v := p.Source.(*dto.Metric)
	return int(formatUint64(v.Summary.SampleCount)), nil
}

func (*summaryMetricImpl) SampleSum(p graphql.ResolveParams) (float64, error) {
	v := p.Source.(*dto.Metric)
	return formatFloat(v.Summary.SampleSum), nil
}

func (*summaryMetricImpl) Quantile(p graphql.ResolveParams) (interface{}, error) {
	v := p.Source.(*dto.Metric)
	return v.Summary.Quantile, nil
}

//
// Implement UntypedMetricFieldResolvers
//

type untypedMetricImpl struct {
	*commonMetricImpl
}

func (*untypedMetricImpl) Value(p graphql.ResolveParams) (float64, error) {
	v := p.Source.(*dto.Metric)
	return formatFloat(v.Untyped.Value), nil
}

//
// Implement HistogramMetricFieldResolvers
//

type histogramMetricImpl struct {
	*commonMetricImpl
}

func (*histogramMetricImpl) SampleCount(p graphql.ResolveParams) (int, error) {
	v := p.Source.(*dto.Metric)
	return int(formatUint64(v.Histogram.SampleCount)), nil
}

func (*histogramMetricImpl) SampleSum(p graphql.ResolveParams) (float64, error) {
	v := p.Source.(*dto.Metric)
	return formatFloat(v.Histogram.SampleSum), nil
}

func (*histogramMetricImpl) Bucket(p graphql.ResolveParams) (interface{}, error) {
	v := p.Source.(*dto.Metric)
	return v.Histogram.Bucket, nil
}

//
// Implement HistogramMetricFieldResolvers
//

type metricFamilyImpl struct {
}

func (*metricFamilyImpl) Name(p graphql.ResolveParams) (string, error) {
	v := p.Source.(*dto.MetricFamily)
	return v.GetName(), nil
}

func (*metricFamilyImpl) Help(p graphql.ResolveParams) (string, error) {
	v := p.Source.(*dto.MetricFamily)
	return v.GetHelp(), nil
}

func (*metricFamilyImpl) Type(p graphql.ResolveParams) (schema.MetricKind, error) {
	v := p.Source.(*dto.MetricFamily)
	return schema.MetricKind(v.Type.String()), nil
}

func (*metricFamilyImpl) Metric(p graphql.ResolveParams) (interface{}, error) {
	v := p.Source.(*dto.MetricFamily)
	return v.GetMetric(), nil
}

type commonMetricImpl struct{}

func (*commonMetricImpl) Labels(p graphql.ResolveParams) (interface{}, error) {
	v := p.Source.(*dto.Metric)
	kv := make([]util_api.KVPairString, 0, len(v.Label))
	for _, pair := range v.Label {
		kv = append(
			kv,
			util_api.KVPairString{
				Key: formatString(pair.Name),
				Val: formatString(pair.Value),
			},
		)
	}
	return kv, nil
}

func (*commonMetricImpl) Timestamp(p graphql.ResolveParams) (*time.Time, error) {
	v := p.Source.(*dto.Metric)
	if ms := v.GetTimestampMs(); ms > 0 {
		ts := unixTimeMs(time.Duration(ms))
		return &ts, nil
	}
	return nil, nil
}

func unixTimeMs(ms time.Duration) time.Time {
	unix := time.Unix(0, 0)
	return unix.Add(ms * time.Millisecond)
}

func formatUint64(f *uint64) uint64 {
	if f == nil {
		return 0
	}
	return *f
}

func formatFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func formatString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
