package statsdaemon

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/backends/sender"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// BackendName is the name of this backend.
	BackendName      = "statsdaemon"
	maxUDPPacketSize = 1472
	maxTCPPacketSize = 1 * 1024 * 1024
	// DefaultDialTimeout is the default net.Dial timeout.
	DefaultDialTimeout = 5 * time.Second
	// DefaultWriteTimeout is the default socket write timeout.
	DefaultWriteTimeout = 30 * time.Second
	// sendChannelSize specifies the size of the buffer of a channel between caller goroutine, producing buffers, and the
	// goroutine that writes them to the socket.
	sendChannelSize = 1000
	// maxConcurrentSends is the number of max concurrent SendMetricsAsync calls that can actually make progress.
	// More calls will block. The current implementation uses maximum 1 call.
	maxConcurrentSends = 10
)

// Client is an object that is used to send messages to a statsd server's UDP or TCP interface.
type Client struct {
	packetSize  int
	disableTags bool
	sender      sender.Sender
}

// overflowHandler is invoked when accumulated packed size has reached it's limit.
// This function should return a new buffer to be used for the rest of the work (may be the same buffer
// if contents are processed somehow and are no longer needed).
type overflowHandler func(*bytes.Buffer) (buf *bytes.Buffer, stop bool)

func (client *Client) Run(ctx context.Context) {
	client.sender.Run(ctx)
}

// SendMetricsAsync flushes the metrics to the statsd server, preparing payload synchronously but doing the send asynchronously.
func (client *Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	sink := make(chan *bytes.Buffer, sendChannelSize)
	select {
	case <-ctx.Done():
		cb([]error{ctx.Err()})
		return
	case client.sender.Sink <- sender.Stream{Ctx: ctx, Cb: cb, Buf: sink}:
	}
	defer close(sink)
	client.processMetrics(metrics, func(buf *bytes.Buffer) (*bytes.Buffer, bool) {
		select {
		case <-ctx.Done():
			return nil, true
		case sink <- buf:
			return client.sender.GetBuffer(), false
		}
	})
}

func (client *Client) processMetrics(metrics *gostatsd.MetricMap, handler overflowHandler) {
	type stopProcessing struct {
	}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(stopProcessing); !ok {
				panic(r)
			}
		}
	}()
	buf := client.sender.GetBuffer()
	defer func() {
		// Have to use a closure because buf pointer might change its value later
		client.sender.PutBuffer(buf)
	}()
	line := new(bytes.Buffer)
	writeLine := func(format, name, tags string, value interface{}) {
		line.Reset()
		if tags == "" || client.disableTags {
			format += "\n"
			fmt.Fprintf(line, format, name, value) // #nosec
		} else {
			format += "|#%s\n"
			fmt.Fprintf(line, format, name, value, tags) // #nosec
		}
		// Make sure we don't go over max udp datagram size
		if buf.Len()+line.Len() > client.packetSize {
			b, stop := handler(buf)
			if stop {
				panic(stopProcessing{})
			}
			buf = b
		}
		fmt.Fprint(buf, line) // #nosec
	}
	metrics.Counters.Each(func(key, tagsKey string, counter gostatsd.Counter) {
		// do not send statsd stats as they will be recalculated on the master instead
		if !strings.HasPrefix(key, "statsd.") {
			writeLine("%s:%d|c", key, tagsKey, counter.Value)
		}
	})
	metrics.Timers.Each(func(key, tagsKey string, timer gostatsd.Timer) {
		for _, tr := range timer.Values {
			writeLine("%s:%f|ms", key, tagsKey, tr)
		}
	})
	metrics.Gauges.Each(func(key, tagsKey string, gauge gostatsd.Gauge) {
		writeLine("%s:%f|g", key, tagsKey, gauge.Value)
	})
	metrics.Sets.Each(func(key, tagsKey string, set gostatsd.Set) {
		for k := range set.Values {
			writeLine("%s:%s|s", key, tagsKey, k)
		}
	})
	if buf.Len() > 0 {
		b, stop := handler(buf) // Process what's left in the buffer
		if !stop {
			buf = b
		}
	}
}

// SendEvent sends events to the statsd master server.
func (client *Client) SendEvent(ctx context.Context, e *gostatsd.Event) error {
	conn, err := client.sender.ConnFactory()
	if err != nil {
		return fmt.Errorf("error connecting to statsd backend: %s", err)
	}
	defer conn.Close()

	_, err = conn.Write(constructEventMessage(e).Bytes())

	return err
}

func constructEventMessage(e *gostatsd.Event) *bytes.Buffer {
	text := strings.Replace(e.Text, "\n", "\\n", -1)

	var buf bytes.Buffer
	buf.WriteString("_e{")
	buf.WriteString(strconv.Itoa(len(e.Title)))
	buf.WriteByte(',')
	buf.WriteString(strconv.Itoa(len(text)))
	buf.WriteString("}:")
	buf.WriteString(e.Title)
	buf.WriteByte('|')
	buf.WriteString(text)

	if e.DateHappened != 0 {
		buf.WriteString("|d:")
		buf.WriteString(strconv.FormatInt(e.DateHappened, 10))
	}
	if e.Hostname != "" {
		buf.WriteString("|h:")
		buf.WriteString(e.Hostname)
	}
	if e.AggregationKey != "" {
		buf.WriteString("|k:")
		buf.WriteString(e.AggregationKey)
	}
	if e.SourceTypeName != "" {
		buf.WriteString("|s:")
		buf.WriteString(e.SourceTypeName)
	}
	if e.Priority != gostatsd.PriNormal {
		buf.WriteString("|p:")
		buf.WriteString(e.Priority.String())
	}
	if e.AlertType != gostatsd.AlertInfo {
		buf.WriteString("|t:")
		buf.WriteString(e.AlertType.String())
	}
	if len(e.Tags) > 0 {
		buf.WriteString("|#")
		buf.WriteString(e.Tags[0])
		for _, tag := range e.Tags[1:] {
			buf.WriteByte(',')
			buf.WriteString(tag)
		}
	}
	return &buf
}

// NewClient constructs a new statsd backend client.
func NewClient(address string, dialTimeout, writeTimeout time.Duration, disableTags, tcpTransport bool, tlsConfig *tls.Config) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("[%s] address is required", BackendName)
	}
	if dialTimeout <= 0 {
		return nil, fmt.Errorf("[%s] dialTimeout should be positive", BackendName)
	}
	if writeTimeout < 0 {
		return nil, fmt.Errorf("[%s] writeTimeout should be non-negative", BackendName)
	}
	log.Infof("[%s] address=%s dialTimeout=%s writeTimeout=%s", BackendName, address, dialTimeout, writeTimeout)
	var packetSize int
	var connFactory func() (net.Conn, error)

	if tlsConfig != nil {
		if !tcpTransport {
			// Avoid surprising a user that expected this to enable DTLS.
			return nil, fmt.Errorf("[%s] tcp_transport is required when using tls_transport", BackendName)
		}

		packetSize = maxTCPPacketSize
		dialer := &net.Dialer{Timeout: dialTimeout}
		connFactory = func() (net.Conn, error) {
			return tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
		}
	} else if tcpTransport {
		packetSize = maxTCPPacketSize
		connFactory = func() (net.Conn, error) {
			return net.DialTimeout("tcp", address, dialTimeout)
		}
	} else {
		packetSize = maxUDPPacketSize
		connFactory = func() (net.Conn, error) {
			return net.DialTimeout("udp", address, dialTimeout)
		}
	}
	return &Client{
		packetSize:  packetSize,
		disableTags: disableTags,
		sender: sender.Sender{
			ConnFactory: connFactory,
			Sink:        make(chan sender.Stream, maxConcurrentSends),
			BufPool: sync.Pool{
				New: func() interface{} {
					buf := new(bytes.Buffer)
					buf.Grow(packetSize)
					return buf
				},
			},
			WriteTimeout: writeTimeout,
		},
	}, nil
}

// NewClientFromViper constructs a statsd client by connecting to an address.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	g := getSubViper(v, "statsdaemon")
	g.SetDefault("dial_timeout", DefaultDialTimeout)
	g.SetDefault("write_timeout", DefaultWriteTimeout)
	g.SetDefault("disable_tags", false)
	g.SetDefault("tcp_transport", false)
	g.SetDefault("tls_transport", false)
	maybeTLSConfig, err := getTLSConfiguration(
		g.GetString("tls_ca_path"),
		g.GetString("tls_cert_path"),
		g.GetString("tls_key_path"),
		g.GetBool("tls_transport"))
	if err != nil {
		return nil, err
	}
	return NewClient(
		g.GetString("address"),
		g.GetDuration("dial_timeout"),
		g.GetDuration("write_timeout"),
		g.GetBool("disable_tags"),
		g.GetBool("tcp_transport"),
		maybeTLSConfig,
	)
}

// Name returns the name of the backend.
func (client *Client) Name() string {
	return BackendName
}

func getSubViper(v *viper.Viper, key string) *viper.Viper {
	n := v.Sub(key)
	if n == nil {
		n = viper.New()
	}
	return n
}
