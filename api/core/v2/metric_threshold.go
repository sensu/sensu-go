package v2

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MetricThresholds []*MetricThreshold

// Validate returns an error if the MetricThreshold does not pass validation tests.
func (m *MetricThreshold) Validate() error {
	if err := ValidateName(m.Name); err != nil {
		return errors.New("metric threshold name " + err.Error())
	}

	if len(m.Thresholds) == 0 {
		return errors.New("there must be at least one threshold rule")
	}

	for i, threshold := range m.Thresholds {
		if err := threshold.Validate(); err != nil {
			return fmt.Errorf("rule %d: %v", i, err)
		}
	}
	for i, tag := range m.Tags {
		if err := tag.Validate(); err != nil {
			return fmt.Errorf("tag %d: %v", i, err)
		}
	}

	return nil
}

func (m *MetricThreshold) MatchesMetricPoint(metric *MetricPoint) bool {
	if m.Name != metric.Name {
		return false
	}
	for _, tag := range m.Tags {
		found := false
		for _, metricTag := range metric.Tags {
			if tag.Name == metricTag.Name && tag.Value == metricTag.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Validate returns an error if the MetricThresholdRule does not pass validation tests.
func (r *MetricThresholdRule) Validate() error {
	r.Min = strings.TrimSpace(r.Min)
	r.Max = strings.TrimSpace(r.Max)
	if r.Min == "" && r.Max == "" {
		return errors.New("One of Min or Max required")
	}
	if r.Min != "" {
		if _, err := strconv.ParseFloat(r.Min, 64); err != nil {
			return errors.New("Min must be a floating point number")
		}
	}
	if r.Max != "" {
		if _, err := strconv.ParseFloat(r.Max, 64); err != nil {
			return errors.New("Max but be a floating point number")
		}
	}

	return nil
}

// Validate returns an error if the MetricThresholdTag does not pass validation tests.
func (t *MetricThresholdTag) Validate() error {
	if err := ValidateName(t.Name); err != nil {
		return errors.New("metric threshold tag name " + err.Error())
	}
	if t.Value == "" {
		return errors.New("metric threshold tag value must not be empty")
	}
	return nil
}

func (mt MetricThresholds) Validate() error {
	for i, threshold := range mt {
		if err := threshold.Validate(); err != nil {
			return fmt.Errorf("invalid metric threshold %d: %s", i, err.Error())
		}
	}
	return nil
}
