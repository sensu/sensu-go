package v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	thresholdNameOnly = &MetricThreshold{Name: "metric-name", Thresholds: []*MetricThresholdRule{thresholdRuleAll}}
	thresholdOneTag   = &MetricThreshold{Name: "metric-name", Thresholds: []*MetricThresholdRule{thresholdRuleAll}, Tags: []*MetricThresholdTag{tag1}}
	thresholdTwoTags  = &MetricThreshold{Name: "metric-name", Thresholds: []*MetricThresholdRule{thresholdRuleAll}, Tags: []*MetricThresholdTag{tag1, tag2}}

	thresholdRuleAll      = &MetricThresholdRule{Min: "3.4", Max: "10.2", Status: 2}
	thresholdRuleNoMin    = &MetricThresholdRule{Min: "", Max: "10.2", Status: 2}
	thresholdRuleNoMax    = &MetricThresholdRule{Min: "", Max: "10.2", Status: 2}
	thresholdRuleNoMinMax = &MetricThresholdRule{Min: "", Max: "", Status: 2}

	tagValid       = &MetricThresholdTag{Name: "tag-name", Value: "tag-value"}
	tagNoName      = &MetricThresholdTag{Name: "", Value: "tag-value"}
	tagInvalidName = &MetricThresholdTag{Name: "#$#$#$", Value: "value"}
	tagNoValue     = &MetricThresholdTag{Name: "tag-name", Value: ""}
	tag1           = &MetricThresholdTag{Name: "tag1", Value: "value1"}
	tag2           = &MetricThresholdTag{Name: "tag2", Value: "value2"}

	metric            = &MetricPoint{Name: "metric-name", Value: 0.1234, Timestamp: time.Now().UnixMilli()}
	metricNoNameMatch = &MetricPoint{Name: "no-name-match", Value: 4.44444, Timestamp: time.Now().UnixMilli()}

	metricTag1            = &MetricTag{Name: "tag1", Value: "value1"}
	metricTag2            = &MetricTag{Name: "tag2", Value: "value2"}
	metricTagNoNameMatch  = &MetricTag{Name: "nomatchname", Value: "value1"}
	metricTagNoValueMatch = &MetricTag{Name: "tag1", Value: "novaluematch"}
)

func TestMetricThresholds_Validate(t *testing.T) {
	tests := []struct {
		name          string
		thresholdName string
		tags          []*MetricThresholdTag
		thresholds    []*MetricThresholdRule
		expectedError string
	}{
		{
			name:          "all ok",
			thresholdName: "thresholdName",
			tags:          nil,
			thresholds:    []*MetricThresholdRule{thresholdRuleAll},
			expectedError: "",
		},
		{
			name:          "invalid name",
			thresholdName: "!@#$%^&*",
			tags:          nil,
			thresholds:    []*MetricThresholdRule{thresholdRuleAll},
			expectedError: "metric threshold name cannot contain spaces or special characters",
		},
		{
			name:          "empty rules",
			thresholdName: "thresholdName",
			tags:          nil,
			thresholds:    []*MetricThresholdRule{},
			expectedError: "there must be at least one threshold rule",
		},
		{
			name:          "nil rules",
			thresholdName: "thresholdName",
			tags:          nil,
			thresholds:    nil,
			expectedError: "there must be at least one threshold rule",
		},
		{
			name:          "invalid rule",
			thresholdName: "thresholdName",
			tags:          nil,
			thresholds:    []*MetricThresholdRule{thresholdRuleAll, thresholdRuleNoMinMax},
			expectedError: "rule 1: One of Min or Max required",
		},
		{
			name:          "invalid tag",
			thresholdName: "thresholdName",
			tags:          []*MetricThresholdTag{tagNoName},
			thresholds:    []*MetricThresholdRule{thresholdRuleAll},
			expectedError: "tag 0: metric threshold tag name must not be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			threshold := MetricThreshold{
				Name:       test.thresholdName,
				Tags:       test.tags,
				Thresholds: test.thresholds,
			}
			err := threshold.Validate()
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}

func TestMetricThreshold_MatchesMetricPoint(t *testing.T) {
	tests := []struct {
		name       string
		threshold  *MetricThreshold
		metric     *MetricPoint
		metricTags []*MetricTag
		matches    bool
	}{
		{
			name:       "threshold no tag - metric no tag",
			threshold:  thresholdNameOnly,
			metric:     metric,
			metricTags: nil,
			matches:    true,
		}, {
			name:       "threshold no tag - metric no tag - no name match",
			threshold:  thresholdNameOnly,
			metric:     metricNoNameMatch,
			metricTags: nil,
			matches:    false,
		}, {
			name:       "threshold no tag - metric one tag",
			threshold:  thresholdNameOnly,
			metric:     metric,
			metricTags: []*MetricTag{metricTag1},
			matches:    true,
		}, {
			name:       "threshold one tag - metric no tag",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: nil,
			matches:    false,
		}, {
			name:       "threshold one tag - metric one tag - match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTag1},
			matches:    true,
		}, {
			name:       "threshold one tag - metric one tag - no name match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTagNoNameMatch},
			matches:    false,
		}, {
			name:       "threshold one tag - metric one tag - no value match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTagNoValueMatch},
			matches:    false,
		}, {
			name:       "threshold one tag - metric two tags - match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTag1, metricTag2},
			matches:    true,
		}, {
			name:       "threshold one tag - metric two tags - no match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTagNoNameMatch, metricTag2},
			matches:    false,
		}, {
			name:       "threshold two tags - metric one tag",
			threshold:  thresholdTwoTags,
			metric:     metric,
			metricTags: []*MetricTag{metricTag1},
			matches:    false,
		}, {
			name:       "threshold two tags - metric two tags - match",
			threshold:  thresholdTwoTags,
			metric:     metric,
			metricTags: []*MetricTag{metricTag1, metricTag2},
			matches:    true,
		}, {
			name:       "threshold two tags - metric two tags - no match",
			threshold:  thresholdOneTag,
			metric:     metric,
			metricTags: []*MetricTag{metricTagNoNameMatch, metricTag2},
			matches:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.metric.Tags = test.metricTags
			matches := test.threshold.MatchesMetricPoint(test.metric)
			assert.Equal(t, test.matches, matches)
		})
	}

}

func TestMetricThresholdRule_Validate(t *testing.T) {
	tests := []struct {
		name          string
		threshold     *MetricThresholdRule
		expectedError string
	}{
		{
			name:          "valid",
			threshold:     thresholdRuleAll,
			expectedError: "",
		}, {
			name:          "no min",
			threshold:     thresholdRuleNoMin,
			expectedError: "",
		}, {
			name:          "no max",
			threshold:     thresholdRuleNoMax,
			expectedError: "",
		}, {
			name:          "no min or max",
			threshold:     thresholdRuleNoMinMax,
			expectedError: "One of Min or Max required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.threshold.Validate()
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}

func TestMetricTresholdTag_Validate(t *testing.T) {
	tests := []struct {
		name          string
		tag           *MetricThresholdTag
		expectedError string
	}{
		{
			name:          "valid",
			tag:           tagValid,
			expectedError: "",
		}, {
			name:          "no name",
			tag:           tagNoName,
			expectedError: "metric threshold tag name must not be empty",
		}, {
			name:          "invalid name",
			tag:           tagInvalidName,
			expectedError: "metric threshold tag name cannot contain spaces or special characters",
		}, {
			name:          "no value",
			tag:           tagNoValue,
			expectedError: "metric threshold tag value must not be empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.tag.Validate()
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}
