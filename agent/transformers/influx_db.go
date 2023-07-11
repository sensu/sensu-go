package transformers

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	v2 "github.com/sensu/core/v2"
	"github.com/sirupsen/logrus"
)

// InfluxList contains a list of Influx values
type InfluxList []Influx

// Influx contains values of influx db line output metric format
type Influx struct {
	Measurement string
	TagSet      []*v2.MetricTag
	FieldSet    []*Field
	Timestamp   int64
}

// Transform transforms a metric in influx db line protocol to Sensu Metric
// Format
func (i InfluxList) Transform() []*v2.MetricPoint {
	var points []*v2.MetricPoint
	for _, influx := range i {
		for _, fieldSet := range influx.FieldSet {
			mp := &v2.MetricPoint{
				Name:      influx.Measurement + "." + fieldSet.Key,
				Value:     fieldSet.Value,
				Timestamp: influx.Timestamp,
				Tags:      influx.TagSet,
			}
			points = append(points, mp)
		}
	}
	return points
}

// ParseInflux parses an influx db line protocol string into an Influx struct
func ParseInflux(event *v2.Event) InfluxList {
	var influxList InfluxList
	fields := logrus.Fields{
		"namespace": event.Check.Namespace,
		"check":     event.Check.Name,
	}

	metric := strings.TrimSpace(event.Check.Output)
	s := bufio.NewScanner(strings.NewReader(metric))
	l := 0

OUTER:
	for s.Scan() {
		line := s.Text()
		fields["line"] = l
		l++
		i := Influx{}
		args := splitWithoutEscaped(line, " ")
		if len(args) != 3 && len(args) != 2 {
			logger.WithFields(fields).WithError(ErrMetricExtraction).Error("influxdb line format requires 2 arguments with a 3rd (optional) timestamp")
			continue
		}

		measurementTag := splitWithoutEscaped(args[0], ",")
		i.Measurement = unescape(measurementTag[0])
		tagList := []*v2.MetricTag{}
		if len(measurementTag) == 1 {
			i.TagSet = tagList
		} else {
			for i, tagSet := range measurementTag {
				if i != 0 {
					ts := splitWithoutEscaped(tagSet, "=")
					if len(ts) != 2 {
						logger.WithFields(fields).WithError(ErrMetricExtraction).Error("metric tag set is invalid, must contain a key=value pair")
						continue OUTER
					}
					tag := &v2.MetricTag{
						Name:  unescape(ts[0]),
						Value: unescape(ts[1]),
					}
					tagList = append(tagList, tag)
				}
			}
			i.TagSet = tagList
		}

		if event.Check.OutputMetricTags != nil {
			i.TagSet = append(i.TagSet, event.Check.OutputMetricTags...)
		}

		fieldSets := splitWithoutEscaped(args[1], ",")
		fieldList := []*Field{}
		for _, fieldSet := range fieldSets {
			fs := splitWithoutEscaped(fieldSet, "=")
			if len(fs) != 2 {
				logger.WithFields(fields).WithError(ErrMetricExtraction).Error("metric field set is invalid, must contain a key=value pair")
				continue OUTER
			}
			f, err := strconv.ParseFloat(fs[1], 64)
			if err != nil {
				logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("metric field value is invalid, must be a float: %s", fs[1])
				continue OUTER
			}
			field := &Field{
				Key:   unescape(fs[0]),
				Value: f,
			}
			fieldList = append(fieldList, field)
		}
		i.FieldSet = fieldList

		if len(args) == 3 {
			timestamp := args[2]
			if len(timestamp) > 10 {
				timestamp = timestamp[:10]
			}
			t, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				logger.WithFields(fields).WithError(ErrMetricExtraction).Errorf("metric timestamp is invalid, third argument must be an int: %s", timestamp)
				continue
			}
			i.Timestamp = t
		} else {
			i.Timestamp = time.Now().UTC().Unix()
		}
		influxList = append(influxList, i)
	}
	if err := s.Err(); err != nil {
		logger.WithFields(fields).WithError(ErrMetricExtraction).Error(err)
	}

	return influxList
}

func splitWithoutEscaped(s, sep string) []string {
	s = strings.ReplaceAll(s, `\`+sep, "\x00")
	segments := strings.Split(s, sep)
	for i := range segments {
		segments[i] = strings.ReplaceAll(segments[i], "\x00", sep)
	}
	return segments
}

func unescape(s string) string {
	s = strings.ReplaceAll(s, `\\`, "\x00")
	s = strings.ReplaceAll(s, `\`, "")
	s = strings.ReplaceAll(s, "\x00", `\`)
	return s
}
