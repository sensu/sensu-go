package graphite

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/atlassian/gostatsd"

	"github.com/ash2k/stager/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreparePayload(t *testing.T) {
	t.Parallel()
	type testData struct {
		config *Config
		result []byte
	}
	metrics := metrics()
	input := []testData{
		{
			config: &Config{
				// Use defaults
			},
			result: []byte("stats_counts.stat1 5 1234\n" +
				"stats.stat1 1.100000 1234\n" +
				"stats.timers.t1.lower 0.000000 1234\n" +
				"stats.timers.t1.upper 0.000000 1234\n" +
				"stats.timers.t1.count 0 1234\n" +
				"stats.timers.t1.count_ps 0.000000 1234\n" +
				"stats.timers.t1.mean 0.000000 1234\n" +
				"stats.timers.t1.median 0.000000 1234\n" +
				"stats.timers.t1.std 0.000000 1234\n" +
				"stats.timers.t1.sum 0.000000 1234\n" +
				"stats.timers.t1.sum_squares 0.000000 1234\n" +
				"stats.timers.t1.count_90 90.000000 1234\n" +
				"stats.gauges.g1 3.000000 1234\n" +
				"stats.sets.users 3 1234\n"),
		},
		{
			config: &Config{
				GlobalPrefix:    addr("gp"),
				PrefixCounter:   addr("pc"),
				PrefixTimer:     addr("pt"),
				PrefixGauge:     addr("pg"),
				PrefixSet:       addr("ps"),
				GlobalSuffix:    addr("gs"),
				LegacyNamespace: addrB(true),
			},
			result: []byte("stats_counts.stat1.gs 5 1234\n" +
				"stats.stat1.gs 1.100000 1234\n" +
				"stats.timers.t1.lower.gs 0.000000 1234\n" +
				"stats.timers.t1.upper.gs 0.000000 1234\n" +
				"stats.timers.t1.count.gs 0 1234\n" +
				"stats.timers.t1.count_ps.gs 0.000000 1234\n" +
				"stats.timers.t1.mean.gs 0.000000 1234\n" +
				"stats.timers.t1.median.gs 0.000000 1234\n" +
				"stats.timers.t1.std.gs 0.000000 1234\n" +
				"stats.timers.t1.sum.gs 0.000000 1234\n" +
				"stats.timers.t1.sum_squares.gs 0.000000 1234\n" +
				"stats.timers.t1.count_90.gs 90.000000 1234\n" +
				"stats.gauges.g1.gs 3.000000 1234\n" +
				"stats.sets.users.gs 3 1234\n"),
		},
		{
			config: &Config{
				GlobalPrefix:    addr("gp"),
				PrefixCounter:   addr("pc"),
				PrefixTimer:     addr("pt"),
				PrefixGauge:     addr("pg"),
				PrefixSet:       addr("ps"),
				GlobalSuffix:    addr("gs"),
				LegacyNamespace: addrB(false),
			},
			result: []byte("gp.pc.stat1.count.gs 5 1234\n" +
				"gp.pc.stat1.rate.gs 1.100000 1234\n" +
				"gp.pt.t1.lower.gs 0.000000 1234\n" +
				"gp.pt.t1.upper.gs 0.000000 1234\n" +
				"gp.pt.t1.count.gs 0 1234\n" +
				"gp.pt.t1.count_ps.gs 0.000000 1234\n" +
				"gp.pt.t1.mean.gs 0.000000 1234\n" +
				"gp.pt.t1.median.gs 0.000000 1234\n" +
				"gp.pt.t1.std.gs 0.000000 1234\n" +
				"gp.pt.t1.sum.gs 0.000000 1234\n" +
				"gp.pt.t1.sum_squares.gs 0.000000 1234\n" +
				"gp.pt.t1.count_90.gs 90.000000 1234\n" +
				"gp.pg.g1.gs 3.000000 1234\n" +
				"gp.ps.users.gs 3 1234\n"),
		},
	}
	for i, td := range input {
		td := td
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			cl, err := NewClient(td.config, gostatsd.TimerSubtypes{})
			require.NoError(t, err)
			b := cl.preparePayload(metrics, time.Unix(1234, 0))
			assert.Equal(t, string(td.result), b.String(), "test %d", i)
		})
	}
}

func TestSendMetricsAsync(t *testing.T) {
	t.Parallel()
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer l.Close()
	addr := l.Addr().String()
	c, err := NewClient(&Config{
		Address: &addr,
	}, gostatsd.TimerSubtypes{})
	require.NoError(t, err)

	var acceptWg sync.WaitGroup
	acceptWg.Add(1)
	go func() {
		defer acceptWg.Done()
		conn, e := l.Accept()
		if !assert.NoError(t, e) {
			return
		}
		defer conn.Close()
		d := make([]byte, 1024)
		for {
			assert.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
			_, e := conn.Read(d)
			if e == io.EOF {
				break
			}
			assert.NoError(t, e)
		}
	}()
	defer acceptWg.Wait()
	defer l.Close()

	var wg wait.Group
	defer wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	wg.StartWithContext(ctx, c.Run)
	var swg sync.WaitGroup
	swg.Add(1)
	c.SendMetricsAsync(ctx, metrics(), func(errs []error) {
		defer swg.Done()
		for i, e := range errs {
			assert.NoError(t, e, i)
		}
	})
	swg.Wait()
}

func metrics() *gostatsd.MetricMap {
	timestamp := gostatsd.Nanotime(time.Unix(123456, 0).UnixNano())

	return &gostatsd.MetricMap{
		Counters: gostatsd.Counters{
			"stat1": map[string]gostatsd.Counter{
				"tag1": {PerSecond: 1.1, Value: 5, Timestamp: timestamp},
			},
		},
		Timers: gostatsd.Timers{
			"t1": map[string]gostatsd.Timer{
				"baz": {
					Values: []float64{10},
					Percentiles: gostatsd.Percentiles{
						gostatsd.Percentile{Float: 90, Str: "count_90"},
					},
					Timestamp: timestamp,
				},
			},
		},
		Gauges: gostatsd.Gauges{
			"g1": map[string]gostatsd.Gauge{
				"baz": {Value: 3, Timestamp: timestamp},
			},
		},
		Sets: gostatsd.Sets{
			"users": map[string]gostatsd.Set{
				"baz": {
					Values: map[string]struct{}{
						"joe":  {},
						"bob":  {},
						"john": {},
					},
					Timestamp: timestamp,
				},
			},
		},
	}
}
