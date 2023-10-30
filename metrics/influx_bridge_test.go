package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	influx "github.com/influxdata/line-protocol"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/assert"
)

const (
	extraLabelName  = "extraLabel"
	extraLabelValue = "extraValue"
)

func newTestGatherer() prometheus.Gatherer {
	collector := collectors.NewGoCollector()
	gatherer := prometheus.NewRegistry()
	gatherer.MustRegister(collector)
	return gatherer
}

func TestInfluxBridgePush(t *testing.T) {
	gatherer := newTestGatherer()
	buf := new(bytes.Buffer)
	cfg := InfluxBridgeConfig{
		Writer:   buf,
		Interval: time.Minute,
		Gatherer: gatherer,
	}
	bridge, err := NewInfluxBridge(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := bridge.Push(); err != nil {
		t.Fatal(err)
	}
	// expect to get at least N metrics
	metrics := strings.Split(buf.String(), "\n")
	if got, want := len(metrics), 34; got < want {
		t.Errorf("expected at least %d metrics, got %d: %+v", want, got, metrics)
	}
	parser := influx.NewStreamParser(buf)
	for {
		metric, err := parser.Next()
		if err == influx.EOF {
			break
		}
		if err != nil {
			t.Error(err)
		}
		if !strings.HasPrefix(metric.Name(), "go") {
			t.Errorf("unexpected metric name: %q", metric.Name())
		}
		if len(metric.FieldList()) == 0 {
			t.Errorf("empty metric field list: %s", metric.Name())
		}
	}
}

func TestInfluxBridgeFilter(t *testing.T) {
	gatherer := newTestGatherer()
	buf := new(bytes.Buffer)
	cfg := InfluxBridgeConfig{
		Writer:   buf,
		Interval: time.Minute,
		Gatherer: gatherer,
		Select:   []string{"go_gc_duration_seconds"},
	}
	bridge, err := NewInfluxBridge(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := bridge.Push(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	parser := influx.NewStreamParser(strings.NewReader(output))
	count := 0
	for {
		metric, err := parser.Next()
		if err == influx.EOF {
			break
		}
		count++
		if err != nil {
			t.Error(err)
		}
		if metric.Name() != "go_gc_duration_seconds" {
			t.Errorf("unexpected metric name: %q", metric.Name())
		}
		if len(metric.FieldList()) == 0 {
			t.Errorf("empty metric field list: %s", metric.Name())
		}
	}
	if got, want := count, 5; got != want {
		t.Errorf("bad count: got %d, want %d", got, want)
		fmt.Println(output)
	}
}

func TestInfluxBridgeExtraLabels(t *testing.T) {
	gatherer := newTestGatherer()
	buf := new(bytes.Buffer)
	cfg := InfluxBridgeConfig{
		Writer:      buf,
		Interval:    time.Minute,
		Gatherer:    gatherer,
		ExtraLabels: map[string]string{extraLabelName: extraLabelValue},
	}
	bridge, err := NewInfluxBridge(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := bridge.Push(); err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	parser := influx.NewStreamParser(strings.NewReader(output))
	count := 0
	for {
		metric, err := parser.Next()
		if err == influx.EOF {
			break
		}
		count++
		if err != nil {
			t.Error(err)
		}
		foundExtraLabel := false
		for _, tag := range metric.TagList() {
			if tag.Key == extraLabelName && tag.Value == extraLabelValue {
				foundExtraLabel = true
				break
			}
		}
		assert.True(t, foundExtraLabel, "extra label not found for metric")
	}
}
