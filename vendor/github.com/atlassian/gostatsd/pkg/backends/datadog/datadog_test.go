package datadog

import (
	"bytes"
	"compress/zlib"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/atlassian/gostatsd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetries(t *testing.T) {
	t.Parallel()
	var requestNum uint32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		n := atomic.AddUint32(&requestNum, 1)
		data, err := ioutil.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotEmpty(t, data)
		if n == 1 {
			// Return error on first request to trigger a retry
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client, err := NewClient(ts.URL, "apiKey123", "tcp", defaultMetricsPerBatch, defaultMaxRequests, true, 1*time.Second, 2*time.Second, 1*time.Second, gostatsd.TimerSubtypes{})
	require.NoError(t, err)
	res := make(chan []error, 1)
	client.SendMetricsAsync(context.Background(), twoCounters(), func(errs []error) {
		res <- errs
	})
	errs := <-res
	for _, err := range errs {
		assert.NoError(t, err)
	}
	assert.EqualValues(t, 2, requestNum)
}

func TestSendMetricsInMultipleBatches(t *testing.T) {
	t.Parallel()
	var requestNum uint32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		atomic.AddUint32(&requestNum, 1)
		data, err := ioutil.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotEmpty(t, data)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client, err := NewClient(ts.URL, "apiKey123", "tcp", 1, defaultMaxRequests, true, 1*time.Second, 2*time.Second, 1*time.Second, gostatsd.TimerSubtypes{})
	require.NoError(t, err)
	res := make(chan []error, 1)
	client.SendMetricsAsync(context.Background(), twoCounters(), func(errs []error) {
		res <- errs
	})
	errs := <-res
	for _, err := range errs {
		assert.NoError(t, err)
	}
	assert.EqualValues(t, 2, requestNum)
}

func TestSendMetrics(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		enc := r.Header.Get("Content-Encoding")
		if enc == "deflate" {
			decompressor, err := zlib.NewReader(bytes.NewReader(data))
			if !assert.NoError(t, err) {
				return
			}
			data, err = ioutil.ReadAll(decompressor)
			assert.NoError(t, err)
		}
		expected := `{"series":[` +
			`{"host":"h1","interval":1.1,"metric":"c1","points":[[100,1.1]],"tags":["tag1"],"type":"rate"},` +
			`{"host":"h1","interval":1.1,"metric":"c1.count","points":[[100,5]],"tags":["tag1"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.lower","points":[[100,0]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.upper","points":[[100,1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.count","points":[[100,1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.count_ps","points":[[100,1.1]],"tags":["tag2"],"type":"rate"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.mean","points":[[100,0.5]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.median","points":[[100,0.5]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.std","points":[[100,0.1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.sum","points":[[100,1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.sum_squares","points":[[100,1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h2","interval":1.1,"metric":"t1.count_90","points":[[100,0.1]],"tags":["tag2"],"type":"gauge"},` +
			`{"host":"h3","interval":1.1,"metric":"g1","points":[[100,3]],"tags":["tag3"],"type":"gauge"},` +
			`{"host":"h4","interval":1.1,"metric":"users","points":[[100,3]],"tags":["tag4"],"type":"gauge"}]}`
		assert.Equal(t, expected, string(data))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cli, err := NewClient(ts.URL, "apiKey123", "tcp", 1000, defaultMaxRequests, true, 1*time.Second, 2*time.Second, 1100*time.Millisecond, gostatsd.TimerSubtypes{})
	require.NoError(t, err)
	cli.now = func() time.Time {
		return time.Unix(100, 0)
	}
	res := make(chan []error, 1)
	cli.SendMetricsAsync(context.Background(), metricsOneOfEach(), func(errs []error) {
		res <- errs
	})
	errs := <-res
	for _, err := range errs {
		assert.NoError(t, err)
	}
}

// twoCounters returns two counters.
func twoCounters() *gostatsd.MetricMap {
	return &gostatsd.MetricMap{
		Counters: gostatsd.Counters{
			"stat1": map[string]gostatsd.Counter{
				"tag1": gostatsd.NewCounter(gostatsd.Nanotime(time.Now().UnixNano()), 5, "", nil),
			},
			"stat2": map[string]gostatsd.Counter{
				"tag2": gostatsd.NewCounter(gostatsd.Nanotime(time.Now().UnixNano()), 50, "", nil),
			},
		},
	}
}

func metricsOneOfEach() *gostatsd.MetricMap {
	return &gostatsd.MetricMap{
		Counters: gostatsd.Counters{
			"c1": map[string]gostatsd.Counter{
				"tag1": {PerSecond: 1.1, Value: 5, Timestamp: gostatsd.Nanotime(100), Hostname: "h1", Tags: gostatsd.Tags{"tag1"}},
			},
		},
		Timers: gostatsd.Timers{
			"t1": map[string]gostatsd.Timer{
				"tag2": {
					Count:      1,
					PerSecond:  1.1,
					Mean:       0.5,
					Median:     0.5,
					Min:        0,
					Max:        1,
					StdDev:     0.1,
					Sum:        1,
					SumSquares: 1,
					Values:     []float64{0, 1},
					Percentiles: gostatsd.Percentiles{
						gostatsd.Percentile{Float: 0.1, Str: "count_90"},
					},
					Timestamp: gostatsd.Nanotime(200),
					Hostname:  "h2",
					Tags:      gostatsd.Tags{"tag2"},
				},
			},
		},
		Gauges: gostatsd.Gauges{
			"g1": map[string]gostatsd.Gauge{
				"tag3": {Value: 3, Timestamp: gostatsd.Nanotime(300), Hostname: "h3", Tags: gostatsd.Tags{"tag3"}},
			},
		},
		Sets: gostatsd.Sets{
			"users": map[string]gostatsd.Set{
				"tag4": {
					Values: map[string]struct{}{
						"joe":  {},
						"bob":  {},
						"john": {},
					},
					Timestamp: gostatsd.Nanotime(400),
					Hostname:  "h4",
					Tags:      gostatsd.Tags{"tag4"},
				},
			},
		},
	}
}
