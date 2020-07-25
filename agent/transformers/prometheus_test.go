package transformers

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/prometheus/common/model"
)

func TestParseProm(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

	testCases := []struct {
		metric           string
		expectedFormat   PromList
		timeInconclusive bool
	}{
		{
			metric: "go_gc_duration_seconds{quantile=\"0\"} 3.3722e-05\n",
			expectedFormat: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			event := types.FixtureEvent("test", "test")
			event.Check.Output = tc.metric
			prom := ParseProm(event)
			if !tc.timeInconclusive {
				assert.Equal(tc.expectedFormat, prom)
			}
		})
	}
}

func TestTransformProm(t *testing.T) {
	assert := assert.New(t)
	ts := time.Now().Unix()

 	testCases := []struct {
 		metric         PromList
 		expectedFormat []*types.MetricPoint
 	}{
 		{
 			metric: PromList{
				&model.Sample{
					Metric: model.Metric{
						model.MetricNameLabel: "go_gc_duration_seconds",
						"quantile":            "0",
					},
					Value:     3.3722e-05,
					Timestamp: model.TimeFromUnix(ts),
				},
 			},
 			expectedFormat: []*types.MetricPoint{
 				{
 					Name:      "go_gc_duration_seconds",
 					Value:     3.3722e-05,
 					Timestamp: ts,
 					Tags: []*types.MetricTag{
 						{
 							Name:  "quantile",
 							Value: "0",
 						},
 					},
 				},
 			},
 		},
 	}

 	for _, tc := range testCases {
 		t.Run("transform", func(t *testing.T) {
 			prom := tc.metric.Transform()
 			assert.Equal(tc.expectedFormat, prom)
 		})
 	}
}
