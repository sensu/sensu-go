package graphql

import (
	"context"
	"reflect"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/graphql"
)

func Test_metricImpl_ResolveType(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		want *graphql.Type
	}{
		{
			name: "counter",
			val:  &dto.Metric{Counter: &dto.Counter{}},
			want: &schema.CounterMetricType,
		},
		{
			name: "gauge",
			val:  &dto.Metric{Gauge: &dto.Gauge{}},
			want: &schema.GaugeMetricType,
		},
		{
			name: "summary",
			val:  &dto.Metric{Summary: &dto.Summary{}},
			want: &schema.SummaryMetricType,
		},
		{
			name: "untyped",
			val:  &dto.Metric{Untyped: &dto.Untyped{}},
			want: &schema.UntypedMetricType,
		},
		{
			name: "histogram",
			val:  &dto.Metric{Histogram: &dto.Histogram{}},
			want: &schema.HistogramMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricImpl{}
			if got := m.ResolveType(tt.val, graphql.ResolveTypeParams{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricImpl.ResolveType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commonMetricImpl_Labels(t *testing.T) {
	nameA := "john"
	nameB := "fred"
	tests := []struct {
		name    string
		val     *dto.Metric
		want    interface{}
		wantErr bool
	}{
		{
			name:    "empty",
			val:     &dto.Metric{Label: []*dto.LabelPair{}},
			want:    []util_api.KVPairString{},
			wantErr: false,
		},
		{
			name: "non-empty",
			val: &dto.Metric{
				Label: []*dto.LabelPair{
					&dto.LabelPair{
						Name:  &nameA,
						Value: &nameB,
					},
				},
			},
			want: []util_api.KVPairString{
				util_api.KVPairString{
					Key: nameA,
					Val: nameB,
				},
			},
			wantErr: false,
		},
		{
			name: "nil values",
			val: &dto.Metric{
				Label: []*dto.LabelPair{
					&dto.LabelPair{
						Name:  nil,
						Value: nil,
					},
				},
			},
			want: []util_api.KVPairString{
				util_api.KVPairString{
					Key: "",
					Val: "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commonMetricImpl{}
			got, err := c.Labels(graphql.ResolveParams{Source: tt.val, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("commonMetricImpl.Labels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("commonMetricImpl.Labels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commonMetricImpl_Timestamp(t *testing.T) {
	var tms int64 = 1000
	var ts time.Time = time.Unix(1, 0)
	tests := []struct {
		name    string
		val     *dto.Metric
		want    *time.Time
		wantErr bool
	}{
		{
			name: "nil",
			val:  &dto.Metric{},
			want: nil,
		},
		{
			name: "one second",
			val: &dto.Metric{
				TimestampMs: &tms,
			},
			want: &ts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commonMetricImpl{}
			got, err := c.Timestamp(graphql.ResolveParams{Source: tt.val, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("commonMetricImpl.Timestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("commonMetricImpl.Timestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metricFamilyImpl_Type(t *testing.T) {
	ekind := dto.MetricType_COUNTER
	tests := []struct {
		name    string
		val     *dto.MetricFamily
		want    schema.MetricKind
		wantErr bool
	}{
		{
			name: "counter",
			val: &dto.MetricFamily{
				Type: &ekind,
			},
			want: "COUNTER",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metricFamilyImpl{}
			got, err := m.Type(graphql.ResolveParams{Source: tt.val, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("metricFamilyImpl.Type() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metricFamilyImpl.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_histogramMetricImpl_Bucket(t *testing.T) {
	tests := []struct {
		name    string
		source  *dto.Metric
		want    interface{}
		wantErr bool
	}{
		{
			name: "returns val",
			source: &dto.Metric{
				Histogram: &dto.Histogram{
					Bucket: []*dto.Bucket{
						&dto.Bucket{},
					},
				},
			},
			want: []*dto.Bucket{
				&dto.Bucket{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &histogramMetricImpl{}
			got, err := h.Bucket(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("histogramMetricImpl.Bucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("histogramMetricImpl.Bucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_summaryMetricImpl_SampleSum(t *testing.T) {
	num := 15.0
	tests := []struct {
		name    string
		source  *dto.Metric
		want    float64
		wantErr bool
	}{
		{
			name: "returns val",
			source: &dto.Metric{
				Summary: &dto.Summary{
					SampleSum: &num,
				},
			},
			want: num,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &summaryMetricImpl{}
			got, err := s.SampleSum(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("summaryMetricImpl.SampleSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("summaryMetricImpl.SampleSum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_histogramMetricImpl_SampleSum(t *testing.T) {
	num := 15.0
	tests := []struct {
		name    string
		source  *dto.Metric
		want    float64
		wantErr bool
	}{
		{
			name: "returns val",
			source: &dto.Metric{
				Histogram: &dto.Histogram{
					SampleSum: &num,
				},
			},
			want: num,
		},
		{
			name: "nil source",
			source: &dto.Metric{
				Histogram: &dto.Histogram{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &histogramMetricImpl{}
			got, err := h.SampleSum(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("histogramMetricImpl.SampleSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("histogramMetricImpl.SampleSum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_summaryMetricImpl_SampleCount(t *testing.T) {
	var num uint64 = 50
	tests := []struct {
		name    string
		source  *dto.Metric
		want    int
		wantErr bool
	}{
		{
			name: "returns val",
			source: &dto.Metric{
				Summary: &dto.Summary{
					SampleCount: &num,
				},
			},
			want: int(num),
		},
		{
			name: "nil source",
			source: &dto.Metric{
				Summary: &dto.Summary{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &summaryMetricImpl{}
			got, err := s.SampleCount(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("summaryMetricImpl.SampleCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("summaryMetricImpl.SampleCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_histogramMetricImpl_SampleCount(t *testing.T) {
	var num uint64 = 50
	tests := []struct {
		name    string
		source  *dto.Metric
		want    int
		wantErr bool
	}{
		{
			name: "returns val",
			source: &dto.Metric{
				Histogram: &dto.Histogram{
					SampleCount: &num,
				},
			},
			want: int(num),
		},
		{
			name: "nil source",
			source: &dto.Metric{
				Histogram: &dto.Histogram{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &histogramMetricImpl{}
			got, err := h.SampleCount(graphql.ResolveParams{Source: tt.source, Context: context.Background()})
			if (err != nil) != tt.wantErr {
				t.Errorf("histogramMetricImpl.SampleCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("histogramMetricImpl.SampleCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
