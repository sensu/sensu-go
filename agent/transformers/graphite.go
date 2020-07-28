package transformers

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// GraphiteList contains a list of Graphite values
type GraphiteList []Graphite

// Graphite contains values of graphite plain text output metric format
type Graphite struct {
	Path      string
	Value     float64
	Timestamp int64
	Tags      []*types.MetricTag
}

// Transform transforms a metric in graphite plain text format to Sensu Metric
// Format
func (g GraphiteList) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, graphite := range g {
		mp := &types.MetricPoint{
			Name:      graphite.Path,
			Value:     graphite.Value,
			Timestamp: graphite.Timestamp,
			Tags:      []*types.MetricTag{},
		}

		if mp.Tags == nil {
			mp.Tags = []*types.MetricTag{}
		}

		points = append(points, mp)
	}
	return points
}

// ParseGraphite parses a graphite plain text string into a Graphite struct
func ParseGraphite(event *types.Event) GraphiteList {
	var graphiteList GraphiteList
	fields := logrus.Fields{
		"namespace": event.Check.Namespace,
		"check":     event.Check.Name,
	}

	metric := strings.TrimSpace(event.Check.Output)
	s := bufio.NewScanner(strings.NewReader(metric))
	l := 0

	for s.Scan() {
		line := s.Text()
		fields["line"] = l
		l++
		g := Graphite{}
		args := strings.Split(line, " ")
		if len(args) != 3 {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Error("graphite plain text format requires exactly 3 arguments")
			continue
		}

		g.Path = args[0]

		f, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("metric value is invalid, second argument must be a float: %s", args[1])
			continue
		}
		g.Value = f

		i, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("metric timestamp is invalid, third argument must be an int: %s", args[2])
			continue
		}
		g.Timestamp = i
		g.Tags = event.Check.OutputMetricTags
		graphiteList = append(graphiteList, g)
	}
	if err := s.Err(); err != nil {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error(err)
	}

	return graphiteList
}
