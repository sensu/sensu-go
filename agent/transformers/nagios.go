package transformers

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// NagiosList contains a list of Nagios metrics
type NagiosList []Nagios

// Nagios contains values of Nagios performance data metric
type Nagios struct {
	Label     string
	Value     float64
	Timestamp int64
	Tags      []*types.MetricTag
}

// Transform transforms a metric in Nagio perfdata format to Sensu Metric Format
func (n NagiosList) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, nagios := range n {
		mp := &types.MetricPoint{
			Name:      nagios.Label,
			Value:     nagios.Value,
			Timestamp: nagios.Timestamp,
			Tags:      nagios.Tags,
		}
		points = append(points, mp)
	}
	return points
}

// ParseNagios parses a Nagios perfdata string into a slice of Nagios struct
func ParseNagios(event *types.Event) NagiosList {
	var nagiosList NagiosList
	fields := logrus.Fields{
		"namespace": event.Check.Namespace,
		"check":     event.Check.Name,
	}

	// Ensure we have some perfdata metrics and not only human-readable text
	output := strings.Split(event.Check.Output, "|")
	if len(output) != 2 {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error("nagios perfdata format requires at least one performance data metric")
		return nagiosList
	}

	// Fetch the perfdata and remove leading & trailing whitespaces
	perfdata := strings.TrimSpace(output[1])

	// Split the perfdata into a slice of metrics
	metrics := strings.Split(perfdata, " ")

	// Create a Nagios metric for each perfdata metrics
	for m, metric := range metrics {
		fields["metric"] = m
		if metric = strings.TrimSpace(metric); len(metric) == 0 {
			// the token was just whitespace, ignore it
			continue
		}
		// Clear everything after ';' and split the label and the value
		parts := strings.Split(metric, ";")
		parts = strings.Split(parts[0], "=")
		if len(parts) != 2 {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("invalid nagios perfdata metric: %q", metric)
			continue
		}

		// Make sure we don't have any whitespace in our label
		label := strings.Replace(parts[0], " ", "_", -1)

		// Remove all non-numeric characters from the value
		re := regexp.MustCompile(`[^-\d\.]`)
		strValue := re.ReplaceAllString(parts[1], "")

		// Parse the value as a float64
		value, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("invalid nagios perfdata metric value: %q", parts[1])
			continue
		}

		// Add this metric to our list
		n := Nagios{
			Label:     label,
			Value:     value,
			Timestamp: event.Check.Executed,
			Tags:      event.Check.OutputMetricTags,
		}
		nagiosList = append(nagiosList, n)
	}

	return nagiosList
}
