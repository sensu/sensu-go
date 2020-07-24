package transformers

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// Transform transforms metrics in the Prometheus Exporter Format to
// the Sensu Metric Format.
func (s model.Vector) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, prom := range s {
                tags := []*types.MetricTag{}
		for ln, lv := range prom.Metric {
			if ln != "__name__" {
				mt := &types.MetricTag{
					Name:  ln,
					Value: lv,
				}
				tags = append(tags, mt)
			}
		}
		n := strings.Replace(prom.Metric["__name__"], "\n", "", -1)
		v := strconv.FormatFloat(float64(prom.Value), 'f', -1, 64)
		mp := &types.MetricPoint{
			Name:      n,
			Value:     v,
			Timestamp: prom.Timestamp,
			Tags:      tags,
		}
		points = append(points, mp)
	}
	return points
}

// ParseProm parses a Prometheus Exporter Format string into an Prometheus Vector (sample).
func ParseProm(event *types.Event) model.Vector {
	fields := logrus.Fields{
		"namespace": event.Check.Namespace,
		"check":     event.Check.Name,
	}

	metricFamilies, err := expfmt.TextParser.TextToMetricFamilies(event.Check.Output)

	if err != nil {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error(err)
	}

	s := model.Vector{}

	decodeOptions := &expfmt.DecodeOptions{
		Timestamp: model.Time(time.Now().Unix()),
	}

	for _, family := range metricFamilies {
		familySamples, _ := expfmt.ExtractSamples(decodeOptions, family)
		s = append(s, familySamples...)
	}

	return s
}
