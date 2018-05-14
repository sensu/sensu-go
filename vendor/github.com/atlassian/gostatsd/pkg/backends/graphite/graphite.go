package graphite

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/backends/sender"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// BackendName is the name of this backend.
	BackendName = "graphite"
	// DefaultAddress is the default address of Graphite server.
	DefaultAddress = "localhost:2003"
	// DefaultDialTimeout is the default net.Dial timeout.
	DefaultDialTimeout = 5 * time.Second
	// DefaultWriteTimeout is the default socket write timeout.
	DefaultWriteTimeout = 30 * time.Second
	// DefaultGlobalPrefix is the default global prefix.
	DefaultGlobalPrefix = "stats"
	// DefaultPrefixCounter is the default counters prefix.
	DefaultPrefixCounter = "counters"
	// DefaultPrefixTimer is the default timers prefix.
	DefaultPrefixTimer = "timers"
	// DefaultPrefixGauge is the default gauges prefix.
	DefaultPrefixGauge = "gauges"
	// DefaultPrefixSet is the default sets prefix.
	DefaultPrefixSet = "sets"
	// DefaultGlobalSuffix is the default global suffix.
	DefaultGlobalSuffix = ""
	// DefaultLegacyNamespace controls whether legacy namespace should be used by default.
	DefaultLegacyNamespace = true
)

const (
	bufSize = 1 * 1024 * 1024
	// maxConcurrentSends is the number of max concurrent SendMetricsAsync calls that can actually make progress.
	// More calls will block. The current implementation uses maximum 1 call.
	maxConcurrentSends = 10
)

var (
	regWhitespace  = regexp.MustCompile(`\s+`)
	regNonAlphaNum = regexp.MustCompile(`[^a-zA-Z\d_.-]`)
)

// Config holds configuration for the Graphite backend.
type Config struct {
	Address         *string
	DialTimeout     *time.Duration
	WriteTimeout    *time.Duration
	GlobalPrefix    *string
	PrefixCounter   *string
	PrefixTimer     *string
	PrefixGauge     *string
	PrefixSet       *string
	GlobalSuffix    *string
	LegacyNamespace *bool
}

// Client is an object that is used to send messages to a Graphite server's TCP interface.
type Client struct {
	sender           sender.Sender
	counterNamespace string
	timerNamespace   string
	gaugesNamespace  string
	setsNamespace    string
	globalSuffix     string
	legacyNamespace  bool
	disabledSubtypes gostatsd.TimerSubtypes
}

func (client *Client) Run(ctx context.Context) {
	client.sender.Run(ctx)
}

// SendMetricsAsync flushes the metrics to the Graphite server, preparing payload synchronously but doing the send asynchronously.
func (client *Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	buf := client.preparePayload(metrics, time.Now())
	sink := make(chan *bytes.Buffer, 1)
	sink <- buf
	close(sink)
	select {
	case <-ctx.Done():
		client.sender.PutBuffer(buf)
		cb([]error{ctx.Err()})
	case client.sender.Sink <- sender.Stream{Ctx: ctx, Cb: cb, Buf: sink}:
	}
}

func (client *Client) preparePayload(metrics *gostatsd.MetricMap, ts time.Time) *bytes.Buffer {
	buf := client.sender.GetBuffer()
	now := ts.Unix()
	if client.legacyNamespace {
		metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
			k := sk(key)
			fmt.Fprintf(buf, "stats_counts.%s%s %d %d\n", k, client.globalSuffix, counter.Value, now)                   // #nosec
			fmt.Fprintf(buf, "%s%s%s %f %d\n", client.counterNamespace, k, client.globalSuffix, counter.PerSecond, now) // #nosec
		})
	} else {
		metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
			k := sk(key)
			fmt.Fprintf(buf, "%s%s.count%s %d %d\n", client.counterNamespace, k, client.globalSuffix, counter.Value, now)    // #nosec
			fmt.Fprintf(buf, "%s%s.rate%s %f %d\n", client.counterNamespace, k, client.globalSuffix, counter.PerSecond, now) // #nosec
		})
	}
	metrics.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		k := sk(key)
		if !client.disabledSubtypes.Lower {
			fmt.Fprintf(buf, "%s%s.lower%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.Min, now) // #nosec
		}
		if !client.disabledSubtypes.Upper {
			fmt.Fprintf(buf, "%s%s.upper%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.Max, now) // #nosec
		}
		if !client.disabledSubtypes.Count {
			fmt.Fprintf(buf, "%s%s.count%s %d %d\n", client.timerNamespace, k, client.globalSuffix, timer.Count, now) // #nosec
		}
		if !client.disabledSubtypes.CountPerSecond {
			fmt.Fprintf(buf, "%s%s.count_ps%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.PerSecond, now) // #nosec
		}
		if !client.disabledSubtypes.Mean {
			fmt.Fprintf(buf, "%s%s.mean%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.Mean, now) // #nosec
		}
		if !client.disabledSubtypes.Median {
			fmt.Fprintf(buf, "%s%s.median%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.Median, now) // #nosec
		}
		if !client.disabledSubtypes.StdDev {
			fmt.Fprintf(buf, "%s%s.std%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.StdDev, now) // #nosec
		}
		if !client.disabledSubtypes.Sum {
			fmt.Fprintf(buf, "%s%s.sum%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.Sum, now) // #nosec
		}
		if !client.disabledSubtypes.SumSquares {
			fmt.Fprintf(buf, "%s%s.sum_squares%s %f %d\n", client.timerNamespace, k, client.globalSuffix, timer.SumSquares, now) // #nosec
		}
		for _, pct := range timer.Percentiles {
			fmt.Fprintf(buf, "%s%s.%s%s %f %d\n", client.timerNamespace, k, pct.Str, client.globalSuffix, pct.Float, now) // #nosec
		}
	})
	metrics.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		fmt.Fprintf(buf, "%s%s%s %f %d\n", client.gaugesNamespace, sk(key), client.globalSuffix, gauge.Value, now) // #nosec
	})
	metrics.Sets.Each(func(key, tagsKey string, set gostatsd.Set) {
		fmt.Fprintf(buf, "%s%s%s %d %d\n", client.setsNamespace, sk(key), client.globalSuffix, len(set.Values), now) // #nosec
	})
	return buf
}

// SendEvent discards events.
func (client *Client) SendEvent(ctx context.Context, e *gostatsd.Event) error {
	return nil
}

// Name returns the name of the backend.
func (client *Client) Name() string {
	return BackendName
}

// NewClientFromViper constructs a GraphiteClient object by connecting to an address.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	g := getSubViper(v, "graphite")
	g.SetDefault("address", DefaultAddress)
	g.SetDefault("dial_timeout", DefaultDialTimeout)
	g.SetDefault("write_timeout", DefaultWriteTimeout)
	g.SetDefault("global_prefix", DefaultGlobalPrefix)
	g.SetDefault("prefix_counter", DefaultPrefixCounter)
	g.SetDefault("prefix_timer", DefaultPrefixTimer)
	g.SetDefault("prefix_gauge", DefaultPrefixGauge)
	g.SetDefault("prefix_set", DefaultPrefixSet)
	g.SetDefault("global_suffix", DefaultGlobalSuffix)
	g.SetDefault("legacy_namespace", DefaultLegacyNamespace)
	return NewClient(&Config{
		Address:         addr(g.GetString("address")),
		DialTimeout:     addrD(g.GetDuration("dial_timeout")),
		WriteTimeout:    addrD(g.GetDuration("write_timeout")),
		GlobalPrefix:    addr(g.GetString("global_prefix")),
		PrefixCounter:   addr(g.GetString("prefix_counter")),
		PrefixTimer:     addr(g.GetString("prefix_timer")),
		PrefixGauge:     addr(g.GetString("prefix_gauge")),
		PrefixSet:       addr(g.GetString("prefix_set")),
		GlobalSuffix:    addr(g.GetString("global_suffix")),
		LegacyNamespace: addrB(g.GetBool("legacy_namespace")),
	}, gostatsd.DisabledSubMetrics(v))
}

// NewClient constructs a Graphite backend object.
func NewClient(config *Config, disabled gostatsd.TimerSubtypes) (*Client, error) {
	address := getOrDefaultStr(config.Address, DefaultAddress)
	if address == "" {
		return nil, fmt.Errorf("[%s] address is required", BackendName)
	}
	dialTimeout := getOrDefaultDur(config.DialTimeout, DefaultDialTimeout)
	if dialTimeout <= 0 {
		return nil, fmt.Errorf("[%s] dialTimeout should be positive", BackendName)
	}
	writeTimeout := getOrDefaultDur(config.WriteTimeout, DefaultWriteTimeout)
	if writeTimeout < 0 {
		return nil, fmt.Errorf("[%s] writeTimeout should be non-negative", BackendName)
	}
	globalSuffix := getOrDefaultStr(config.GlobalSuffix, DefaultGlobalSuffix)
	if globalSuffix != "" {
		globalSuffix = `.` + globalSuffix
	}
	var counterNamespace, timerNamespace, gaugesNamespace, setsNamespace string
	var legacyNamespace bool
	if config.LegacyNamespace != nil {
		legacyNamespace = *config.LegacyNamespace
	} else {
		legacyNamespace = DefaultLegacyNamespace
	}
	if legacyNamespace {
		counterNamespace = DefaultGlobalPrefix + `.`
		timerNamespace = DefaultGlobalPrefix + ".timers."
		gaugesNamespace = DefaultGlobalPrefix + ".gauges."
		setsNamespace = DefaultGlobalPrefix + ".sets."
	} else {
		globalPrefix := getOrDefaultPrefix(config.GlobalPrefix, DefaultGlobalPrefix)
		counterNamespace = globalPrefix + getOrDefaultPrefix(config.PrefixCounter, DefaultPrefixCounter)
		timerNamespace = globalPrefix + getOrDefaultPrefix(config.PrefixTimer, DefaultPrefixTimer)
		gaugesNamespace = globalPrefix + getOrDefaultPrefix(config.PrefixGauge, DefaultPrefixGauge)
		setsNamespace = globalPrefix + getOrDefaultPrefix(config.PrefixSet, DefaultPrefixSet)
	}
	log.Infof("[%s] address=%s dialTimeout=%s writeTimeout=%s", BackendName, address, dialTimeout, writeTimeout)
	return &Client{
		sender: sender.Sender{
			ConnFactory: func() (net.Conn, error) {
				return net.DialTimeout("tcp", address, dialTimeout)
			},
			Sink: make(chan sender.Stream, maxConcurrentSends),
			BufPool: sync.Pool{
				New: func() interface{} {
					buf := new(bytes.Buffer)
					buf.Grow(bufSize)
					return buf
				},
			},
			WriteTimeout: writeTimeout,
		},
		counterNamespace: counterNamespace,
		timerNamespace:   timerNamespace,
		gaugesNamespace:  gaugesNamespace,
		setsNamespace:    setsNamespace,
		globalSuffix:     globalSuffix,
		legacyNamespace:  legacyNamespace,
		disabledSubtypes: disabled,
	}, nil
}

func getOrDefaultPrefix(val *string, def string) string {
	v := getOrDefaultStr(val, def)
	if v != "" {
		return v + `.`
	}
	return ""
}

func getOrDefaultStr(val *string, def string) string {
	if val != nil {
		return *val
	}
	return def
}

func getOrDefaultDur(val *time.Duration, def time.Duration) time.Duration {
	if val != nil {
		return *val
	}
	return def
}

func sk(s string) []byte {
	r1 := regWhitespace.ReplaceAllLiteral([]byte(s), []byte{'_'})
	r2 := bytes.Replace(r1, []byte{'/'}, []byte{'-'}, -1)
	return regNonAlphaNum.ReplaceAllLiteral(r2, nil)
}

func addr(s string) *string {
	return &s
}

func addrB(b bool) *bool {
	return &b
}

func addrD(d time.Duration) *time.Duration {
	return &d
}

func getSubViper(v *viper.Viper, key string) *viper.Viper {
	n := v.Sub(key)
	if n == nil {
		n = viper.New()
	}
	return n
}
