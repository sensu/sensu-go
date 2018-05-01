package transformers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// GraphiteList contains a list of Graphite values
type GraphiteList []Graphite

// Graphite contains values of graphite plain text metric format
type Graphite struct {
	Path      string
	Value     float64
	Timestamp int64
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
		points = append(points, mp)
	}
	return points
}

// ParseGraphite parses a graphite plain text string into a Graphite struct
func ParseGraphite(metric string) (GraphiteList, error) {
	var graphites GraphiteList
	lines := strings.Split(metric, "\n")
	for _, line := range lines {
		g := Graphite{}
		args := strings.Split(line, " ")
		if len(args) != 3 {
			return []Graphite{}, errors.New("graphite plain text format requires exactly 3 arguments")
		}

		g.Path = args[0]

		f, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return []Graphite{}, errors.New("metric value is invalid, second argument must be a float")
		}
		g.Value = f

		i, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return []Graphite{}, errors.New("metric timestamp is invalid, third argument must be an int")
		}
		g.Timestamp = i
		graphites = append(graphites, g)
	}

	return graphites, nil
}
