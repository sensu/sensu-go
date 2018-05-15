package null

import (
	"context"

	"github.com/atlassian/gostatsd"

	"github.com/spf13/viper"
)

// BackendName is the name of this backend.
const BackendName = "null"

// Client represents a discarding backend.
type Client struct{}

// NewClientFromViper constructs a GraphiteClient object by connecting to an address.
func NewClientFromViper(v *viper.Viper) (gostatsd.Backend, error) {
	return NewClient()
}

// NewClient constructs a client object.
func NewClient() (*Client, error) {
	return &Client{}, nil
}

// SendMetricsAsync discards the metrics in a MetricsMap.
func (Client) SendMetricsAsync(ctx context.Context, metrics *gostatsd.MetricMap, cb gostatsd.SendCallback) {
	cb(nil)
}

// SendEvent discards events.
func (Client) SendEvent(ctx context.Context, e *gostatsd.Event) error {
	return nil
}

// Name returns the name of the backend.
func (Client) Name() string {
	return BackendName
}
