package transformers

import (
	"math"
	"strings"

	time "github.com/echlebek/timeproxy"
	v2 "github.com/sensu/core/v2"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

const (
	PromTypeTagName	= "prom_type"
	PromHelpTagName	= "prom_help"
)

// PromList contains Prometheus vector (samples)
type PromList model.Vector

// Transform transforms metrics in the Prometheus Exposition Text
// Format to the Sensu Metric Format.
func (p PromList) Transform() []*v2.MetricPoint {
	var points []*v2.MetricPoint
	for _, prom := range p {
		v := float64(prom.Value)
		if math.IsNaN(v) {
			continue
		}
		tags := []*v2.MetricTag{}
		for ln, lv := range prom.Metric {
			if ln != "__name__" {
				mt := &v2.MetricTag{
					Name:	string(ln),
					Value:	string(lv),
				}
				tags = append(tags, mt)
			}
		}
		n := string(prom.Metric["__name__"])
		n = strings.Replace(n, "\n", "", -1)
		mp := &v2.MetricPoint{
			Name:		n,
			Value:		v,
			Timestamp:	prom.Timestamp.Unix(),
			Tags:		tags,
		}
		points = append(points, mp)
	}
	return points
}

// ParseProm parses a Prometheus Exposition Text Formated string into
// an Prometheus Vector (sample).
func ParseProm(event *v2.Event) PromList {
	fields := logrus.Fields{
		"namespace":	event.Check.Namespace,
		"check":	event.Check.Name,
	}

	t := strings.NewReader(event.Check.Output)
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(t)

	if err != nil {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error(err)
	}

	p := PromList{}

	decodeOptions := &expfmt.DecodeOptions{
		Timestamp: model.TimeFromUnix(time.Now().Unix()),
	}

	for _, family := range metricFamilies {
		familySamples, _ := expfmt.ExtractSamples(decodeOptions, family)
		for _, prom := range familySamples {
			lv := model.LabelValue(strings.ToLower(family.Type.String()))
			prom.Metric[model.LabelName(PromTypeTagName)] = lv
			if help := family.GetHelp(); help != "" {
				prom.Metric[model.LabelName(PromHelpTagName)] = model.LabelValue(help)
			}
		}
		p = append(p, familySamples...)
	}

	if len(event.Check.OutputMetricTags) > 0 {
		for _, prom := range p {
			for _, tag := range event.Check.OutputMetricTags {
				prom.Metric[model.LabelName(tag.Name)] = model.LabelValue(tag.Value)
			}
		}
	}

	return p
}
