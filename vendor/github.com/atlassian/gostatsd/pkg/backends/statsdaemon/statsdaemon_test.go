package statsdaemon

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/atlassian/gostatsd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var longName = strings.Repeat("t", maxUDPPacketSize-5)
var m = gostatsd.MetricMap{
	Counters: gostatsd.Counters{
		longName: map[string]gostatsd.Counter{
			"tag1": gostatsd.NewCounter(gostatsd.Nanotime(time.Now().UnixNano()), 5, "", nil),
		},
	},
}

func TestProcessMetricsRecover(t *testing.T) {
	t.Parallel()
	c, err := NewClient("localhost:8125", 1*time.Second, 1*time.Second, false, false, nil)
	require.NoError(t, err)
	c.processMetrics(&m, func(buf *bytes.Buffer) (*bytes.Buffer, bool) {
		return nil, true
	})
}

func TestProcessMetricsPanic(t *testing.T) {
	t.Parallel()
	c, err := NewClient("localhost:8125", 1*time.Second, 1*time.Second, false, false, nil)
	require.NoError(t, err)
	expectedErr := errors.New("ABC some error")
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, expectedErr, r)
		} else {
			t.Error("should have panicked")
		}
	}()
	c.processMetrics(&m, func(buf *bytes.Buffer) (*bytes.Buffer, bool) {
		panic(expectedErr)
	})
	t.Error("unreachable")
}

var gaugeMetic = gostatsd.MetricMap{
	Gauges: gostatsd.Gauges{
		"statsd.processing_time": map[string]gostatsd.Gauge{
			"tag1": gostatsd.NewGauge(gostatsd.Nanotime(time.Now().UnixNano()), 2, "", nil),
		},
	},
}

func TestProcessMetrics(t *testing.T) {
	t.Parallel()
	input := []struct {
		disableTags   bool
		expectedValue string
	}{
		{
			disableTags:   false,
			expectedValue: "statsd.processing_time:2.000000|g|#tag1\n",
		},
		{
			disableTags:   true,
			expectedValue: "statsd.processing_time:2.000000|g\n",
		},
	}
	for _, val := range input {
		val := val
		t.Run(fmt.Sprintf("disableTags: %t", val.disableTags), func(t *testing.T) {
			t.Parallel()
			c, err := NewClient("localhost:8125", 1*time.Second, 1*time.Second, val.disableTags, false, nil)
			require.NoError(t, err)
			c.processMetrics(&gaugeMetic, func(buf *bytes.Buffer) (*bytes.Buffer, bool) {
				assert.EqualValues(t, val.expectedValue, buf.String())
				return new(bytes.Buffer), false
			})
		})
	}
}
