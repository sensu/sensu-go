package transformers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// InfluxList contains a list of Influx values
type InfluxList []Influx

// Influx contains values of influx db line metric format
type Influx struct {
	Measurement string
	TagSet      []Tag
	FieldSet    []Field
	Timestamp   int64
}

// Tag is a key value pair representing an influxdb tag set
type Tag struct {
	Key   string
	Value string
}

// Field is a key value pair representing an influxdb field set
type Field struct {
	Key   string
	Value float64
}

// Transform transforms a metric in influx db line protocol to Sensu Metric
// Format
func (i InfluxList) Transform() []*types.MetricPoint {
	var points []*types.MetricPoint
	for _, influx := range i {
		for _, fieldSet := range influx.FieldSet {
			mp := &types.MetricPoint{
				Name:      influx.Measurement + "." + fieldSet.Key,
				Value:     fieldSet.Value,
				Timestamp: influx.Timestamp,
				Tags:      tagSetToMetricTag(influx.TagSet),
			}
			points = append(points, mp)
		}
	}
	return points
}

func tagSetToMetricTag(tagSet []Tag) []*types.MetricTag {
	tags := []*types.MetricTag{}
	for _, tag := range tagSet {
		t := &types.MetricTag{
			Name:  tag.Key,
			Value: tag.Value,
		}
		tags = append(tags, t)
	}
	return tags
}

// ParseInflux parses an influx db line protocol string into an Influx struct
func ParseInflux(metric string) (InfluxList, error) {
	var influxList InfluxList
	metric = strings.TrimSpace(metric)
	lines := strings.Split(metric, "\n")
	for _, line := range lines {
		i := Influx{}
		args := strings.Split(line, " ")
		if len(args) != 3 {
			return InfluxList{}, errors.New("influxdb line format requires exactly 3 arguments")
		}

		measurementTag := strings.Split(args[0], ",")
		i.Measurement = measurementTag[0]
		tagList := []Tag{}
		if len(measurementTag) == 1 {
			i.TagSet = tagList
		} else {
			for i, tagSet := range measurementTag {
				if i != 0 {
					ts := strings.Split(tagSet, "=")
					if len(ts) != 2 {
						return InfluxList{}, errors.New("metric tag set is invalid, must contain a key=value pair")
					}
					tag := Tag{
						Key:   ts[0],
						Value: ts[1],
					}
					tagList = append(tagList, tag)
				}
			}
			i.TagSet = tagList
		}

		fieldSets := strings.Split(args[1], ",")
		fieldList := []Field{}
		for _, fieldSet := range fieldSets {
			fs := strings.Split(fieldSet, "=")
			if len(fs) != 2 {
				return InfluxList{}, errors.New("metric field set is invalid, must contain a key=value pair")
			}
			f, err := strconv.ParseFloat(fs[1], 64)
			if err != nil {
				return InfluxList{}, errors.New("metric field value is invalid, must be a float")
			}
			field := Field{
				Key:   fs[0],
				Value: f,
			}
			fieldList = append(fieldList, field)
		}
		i.FieldSet = fieldList

		t, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return InfluxList{}, errors.New("metric timestamp is invalid, third argument must be an int")
		}
		i.Timestamp = t
		influxList = append(influxList, i)
	}

	return influxList, nil
}
