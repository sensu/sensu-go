package transformers

import (
	"strconv"
	"strings"
	"time"
	"fmt"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// PromList contains Prometheus vectors/samples
type PromList model.Vector

// Transform transforms metrics in the Prometheus Exporter Format to
// the Sensu Metric Format.
func (p PromList) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, prom := range p {
                tags := []*types.MetricTag{}
		for ln, lv := range prom.Metric {
			if ln != "__name__" {
				mt := &types.MetricTag{
					Name:  fmt.Sprintf("%s", ln),
					Value: fmt.Sprintf("%s", lv),
				}
				tags = append(tags, mt)
			}
		}
		n := fmt.Sprintf("%s", prom.Metric["__name__"])
		n = strings.Replace(n, "\n", "", -1)
		v, _ := strconv.ParseFloat(fmt.Sprintf("%f", prom.Value), 32)
		mp := &types.MetricPoint{
			Name:      n,
			Value:     v,
			Timestamp: prom.Timestamp.Unix(),
			Tags:      tags,
		}
		points = append(points, mp)
	}
	return points
}

// ParseProm parses a Prometheus Exporter Format string into an Prometheus Vector (sample).
func ParseProm(event *types.Event) PromList {
	fields := logrus.Fields{
		"namespace": event.Check.Namespace,
		"check":     event.Check.Name,
	}

	t := strings.NewReader(event.Check.Output)
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(t)

	if err != nil {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error(err)
	}

	p := PromList{}

	decodeOptions := &expfmt.DecodeOptions{
		Timestamp: model.Time(time.Now().Unix()),
	}

	for _, family := range metricFamilies {
		familySamples, _ := expfmt.ExtractSamples(decodeOptions, family)
		p = append(p, familySamples...)
	}

	return p
}
