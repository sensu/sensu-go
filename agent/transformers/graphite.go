package transformers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// Graphite contains values of graphite plain text metric format
type Graphite struct {
	Path      string
	Value     float64
	Timestamp int64
}

// Transform transforms a metric in graphite plain text format to Sensu Metric
// Format
func (g Graphite) Transform() []*types.MetricPoint {
	mp := &types.MetricPoint{
		Name:      g.Path,
		Value:     g.Value,
		Timestamp: g.Timestamp,
		Tags:      []*types.MetricTag{},
	}
	return []*types.MetricPoint{mp}
}

// ParseGraphite parses a graphite plain text string into a Graphite struct
func ParseGraphite(metric string) (Graphite, error) {
	g := Graphite{}
	args := strings.Split(metric, " ")
	if len(args) != 3 {
		return Graphite{}, errors.New("graphite plain text format requires exactly 3 arguments")
	}

	g.Path = args[0]

	f, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return Graphite{}, errors.New("metric value is invalid, second argument must be a float")
	}
	g.Value = f

	i, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return Graphite{}, errors.New("metric timestamp is invalid, third argument must be an int")
	}
	g.Timestamp = i

	return g, nil
}
